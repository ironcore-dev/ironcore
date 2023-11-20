// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package sync_test

import (
	"context"
	"fmt"
	"sync"

	. "github.com/ironcore-dev/ironcore/broker/common/sync"
	. "github.com/ironcore-dev/ironcore/utils/testing"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sync", func() {
	ctx := SetupContext()

	Context("MutexMap", func() {
		const (
			key1 = "foo"
			key2 = "bar"
		)

		It("should lock by key and correctly garbage collect", MustPassRepeatedly(10), func() {
			m := NewMutexMap[string]()

			key1LR1 := NewLockRunner(ctx, m.Locker(key1))
			key1LR2 := NewLockRunner(ctx, m.Locker(key1))

			By("making lock runner 1 obtain key 1")
			key1LR1.Lock()

			By("waiting for the map and key count to be updated")
			Eventually(func(g Gomega) {
				g.Expect(m.Len()).To(Equal(1))
				g.Expect(m.Count(key1)).To(Equal(1))
			}).Should(Succeed())

			By("making lock runner 2 lock key 1")
			key1LR2.Lock()

			By("waiting for the lock count for key 1 to become 2")
			Eventually(func() int { return m.Count(key1) }).Should(Equal(2))

			By("asserting lock runner 2 cannot obtain the key 1")
			Consistently(key1LR2.HoldingLock).Should(BeFalse(), "lock runner 1 should not hold lock")

			By("making lock runner 1 unlock key 1")
			key1LR1.Unlock()

			By("waiting for lock runner 2 to obtain key 1")
			Eventually(key1LR2.HoldingLock).Should(BeTrue(), "lock runner 1 should hold lock")

			By("asserting the lock count to be 1")
			Expect(m.Count(key1)).To(Equal(1))

			By("making lock runner 2 release lock key 1")
			key1LR2.Unlock()

			By("waiting for the map to be empty again")
			Eventually(m.Len).Should(BeZero(), "map should be empty")
		})

		It("should allow multiple locks with different keys to be obtained simultaneously", MustPassRepeatedly(10), func() {
			m := NewMutexMap[string]()

			key1LR := NewLockRunner(ctx, m.Locker(key1))
			key2LR := NewLockRunner(ctx, m.Locker(key2))

			By("making key 1 lock runner lock")
			key1LR.Lock()

			By("making key 2 lock runner lock")
			key2LR.Lock()

			By("waiting for both lock runners to simultaneously hold keys")
			Eventually(func(g Gomega) {
				g.Expect(m.Count(key1)).To(Equal(1))
				g.Expect(m.Count(key2)).To(Equal(1))
				g.Expect(m.Len()).To(Equal(2))
				g.Expect(key1LR.HoldingLock()).To(BeTrue(), "key 1 lock runner should hold lock")
				g.Expect(key2LR.HoldingLock()).To(BeTrue(), "key 2 lock runner should hold lock")
			}).Should(Succeed())
		})
	})
})

type lockRunnerCommand uint8

const (
	lockRunnerCommandLock = iota
	lockRunnerCommandUnlock
)

type LockRunner struct {
	mu sync.RWMutex

	holdingLock bool

	locker   sync.Locker
	commands chan lockRunnerCommand
}

func NewLockRunner(ctx context.Context, locker sync.Locker) *LockRunner {
	lr := &LockRunner{
		locker:   locker,
		commands: make(chan lockRunnerCommand),
	}
	go lr.loop(ctx)
	return lr
}

func (r *LockRunner) loop(ctx context.Context) {
	defer close(r.commands)
	for {
		select {
		case <-ctx.Done():
			return
		case command := <-r.commands:
			switch command {
			case lockRunnerCommandLock:
				r.locker.Lock()
				r.mu.Lock()
				r.holdingLock = true
				r.mu.Unlock()
			case lockRunnerCommandUnlock:
				r.locker.Unlock()
				r.mu.Lock()
				r.holdingLock = false
				r.mu.Unlock()
			default:
				panic(fmt.Errorf("unknown command %d", command))
			}
		}
	}
}

func (r *LockRunner) Lock() {
	r.commands <- lockRunnerCommandLock
}

func (r *LockRunner) Unlock() {
	r.commands <- lockRunnerCommandUnlock
}

func (r *LockRunner) HoldingLock() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.holdingLock
}
