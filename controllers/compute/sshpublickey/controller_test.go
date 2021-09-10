package sshpublickey

import (
	"context"
	_ "embed"
	common "github.com/onmetal/onmetal-api/apis/common/v1alpha1"
	computev1alpha1 "github.com/onmetal/onmetal-api/apis/compute/v1alpha1"
	. "github.com/onmetal/onmetal-api/pkg/utils/testutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var (
	//go:embed testdata/id_rsa.pub
	testPublicKeyData string
)

var _ = Describe("SSHPublicKeyController", func() {
	var (
		namespace     string
		configMapName string
		configMap     *corev1.ConfigMap
		configMapKey  client.ObjectKey

		publicKeyName string
		publicKeyKey  client.ObjectKey
		publicKey     *computev1alpha1.SSHPublicKey
	)
	BeforeEach(func() {
		namespace = corev1.NamespaceDefault

		configMapName = "my-public-key-cm"
		configMapKey = client.ObjectKey{Namespace: namespace, Name: configMapName}
		configMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      configMapName,
			},
			Data: map[string]string{
				computev1alpha1.DefaultSSHPublicKeyDataKey: testPublicKeyData,
			},
		}

		publicKeyName = "my-public-key"
		publicKeyKey = client.ObjectKey{Namespace: namespace, Name: publicKeyName}
		publicKey = &computev1alpha1.SSHPublicKey{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: namespace,
				Name:      publicKeyName,
			},
			Spec: computev1alpha1.SSHPublicKeySpec{
				ConfigMapRef: common.ConfigMapKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
				},
			},
		}
	})
	AfterEach(func() {
		ctx := context.Background()
		for _, toDelete := range []struct {
			key client.ObjectKey
			obj client.Object
		}{
			{
				key: configMapKey,
				obj: configMap,
			},
			{
				key: publicKeyKey,
				obj: publicKey,
			},
		} {
			toDelete.obj.SetNamespace(toDelete.key.Namespace)
			toDelete.obj.SetName(toDelete.key.Name)
			Expect(client.IgnoreNotFound(k8sClient.Delete(ctx, toDelete.obj))).To(Succeed())
		}
	})

	It("should parse and update the key fields", func() {
		ctx := context.Background()
		startTime := metav1.NewTime(time.Now().Add(-1 * time.Second).Round(time.Second))
		By("creating a config map containing an ssh public key")
		Expect(k8sClient.Create(ctx, configMap)).To(Succeed())

		By("creating a public key referencing that config map")
		Expect(k8sClient.Create(ctx, publicKey)).To(Succeed())

		By("waiting for the status to report the key as available")
		Eventually(func() *computev1alpha1.SSHPublicKey {
			Expect(k8sClient.Get(ctx, publicKeyKey, publicKey)).To(Succeed())
			return publicKey
		}, 10*time.Second).Should(WithTransform(func(publicKey *computev1alpha1.SSHPublicKey) []computev1alpha1.SSHPublicKeyCondition {
			return publicKey.Status.Conditions
		}, ContainElement(MatchFields(IgnoreMissing|IgnoreExtras, Fields{
			"Type":               Equal(computev1alpha1.SSHPublicKeyAvailable),
			"Status":             Equal(corev1.ConditionTrue),
			"Reason":             Equal("Valid"),
			"Message":            Equal("The key is well-formed."),
			"LastUpdateTime":     BeMetaV1Temporally(">=", startTime),
			"LastTransitionTime": BeMetaV1Temporally(">=", startTime),
			"ObservedGeneration": Equal(publicKey.Generation),
		}))))
	})
})
