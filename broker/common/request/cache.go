// Copyright 2023 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package request

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/onmetal/onmetal-api/utils/container/list"
	"github.com/onmetal/onmetal-api/utils/generic"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	DefaultCacheTTL           time.Duration = 1 * time.Minute
	DefaultCacheTokenMaxTries int           = 10
	DefaultCacheTokenLen      int           = 8
	DefaultCacheMaxInFlight   int           = 1000
)

type cache[E any] struct {
	mu sync.Mutex

	tokens          map[string]*list.Element[*cacheEntry[E]]
	tokensByAgeDesc *list.List[*cacheEntry[E]]

	ttl           time.Duration
	maxInFlight   int
	tokenMaxTries int
	tokenLen      int
}

type cacheEntry[E any] struct {
	token      string
	request    E
	expiration time.Time
}

type CacheOptions struct {
	TTL           time.Duration
	MaxInFlight   int
	TokenMaxTries int
	TokenLen      int
}

// NewCache constructs a new Cache that tracks request tokens with requests of type E.
func NewCache[E any](opts ...func(*CacheOptions)) Cache[E] {
	o := &CacheOptions{}
	for _, opt := range opts {
		opt(o)
	}

	if o.TTL <= 0 {
		o.TTL = DefaultCacheTTL
	}
	if o.MaxInFlight <= 0 {
		o.MaxInFlight = DefaultCacheMaxInFlight
	}
	if o.TokenMaxTries <= 0 {
		o.TokenMaxTries = DefaultCacheTokenMaxTries
	}
	if o.TokenLen <= 0 {
		o.TokenLen = DefaultCacheTokenLen
	}

	return &cache[E]{
		tokens:          make(map[string]*list.Element[*cacheEntry[E]]),
		tokensByAgeDesc: list.New[*cacheEntry[E]](),
		ttl:             o.TTL,
		maxInFlight:     o.MaxInFlight,
		tokenMaxTries:   o.TokenMaxTries,
		tokenLen:        o.TokenLen,
	}
}

// Insert implements Cache.
func (c *cache[E]) Insert(req E) (token string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.gc()

	if c.tokensByAgeDesc.Len() == c.maxInFlight {
		return "", status.Error(codes.ResourceExhausted, "maximum number of in-flight requests exceeded")
	}

	token, err = c.uniqueToken()
	if err != nil {
		return "", err
	}

	elem := c.tokensByAgeDesc.PushFront(&cacheEntry[E]{token, req, time.Now().Add(c.ttl)})
	c.tokens[token] = elem
	return token, nil
}

// Consume implements Cache.
func (c *cache[E]) Consume(token string) (req E, found bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	elem, ok := c.tokens[token]
	if !ok {
		return generic.Zero[E](), false
	}

	c.tokensByAgeDesc.Remove(elem)
	delete(c.tokens, token)

	entry := elem.Value
	if time.Now().After(entry.expiration) {
		return generic.Zero[E](), false
	}
	return entry.request, true
}

func (c *cache[E]) uniqueToken() (string, error) {
	tokenSize := math.Ceil(float64(c.tokenLen) * 6 / 8)
	rawToken := make([]byte, int(tokenSize))
	for i := 0; i < c.tokenMaxTries; i++ {
		if _, err := rand.Read(rawToken); err != nil {
			return "", err
		}

		encoded := base64.RawURLEncoding.EncodeToString(rawToken)
		token := encoded[:c.tokenLen]

		if _, exists := c.tokens[encoded]; !exists {
			return token, nil
		}
	}
	return "", fmt.Errorf("failed to generate unique token")
}

func (c *cache[E]) gc() {
	now := time.Now()
	for c.tokensByAgeDesc.Len() > 0 {
		oldest := c.tokensByAgeDesc.Back()
		entry := oldest.Value
		if !now.After(entry.expiration) {
			return
		}

		c.tokensByAgeDesc.Remove(oldest)
		delete(c.tokens, entry.token)
	}
}

// Cache associates requests of type E with request tokens. This allows requests to be approved and later on be
// picked up with their associated token.
type Cache[E any] interface {
	// Insert creates a new unique token for the request object and returns it. If it's not possible to generate
	// a unique token, an error is returned.
	Insert(req E) (token string, err error)
	// Consume looks up the cache for the request associated with the given token and returns it if found.
	// If the token is unknown or expired, no request & found = false is returned.
	Consume(token string) (req E, found bool)
}
