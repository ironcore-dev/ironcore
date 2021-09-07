package condition_test

import (
	"github.com/onmetal/onmetal-api/pkg/utils/condition"
	. "github.com/onmetal/onmetal-api/pkg/utils/testutils"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/clock"
	"time"
)

var _ = Describe("Condition", func() {
	type TimePtrCondition struct {
		Type               string
		Status             corev1.ConditionStatus
		LastUpdateTime     *metav1.Time
		LastTransitionTime *metav1.Time
		Reason             string
		Message            string
	}

	type NoTimeCondition struct {
		Type    string
		Status  corev1.ConditionStatus
		Reason  string
		Message string
	}

	Context("Accessor", func() {
		var (
			now     time.Time
			metaNow metav1.Time
			acc     *condition.Accessor

			deployCond  appsv1.DeploymentCondition
			deployConds []appsv1.DeploymentCondition
			custCond    TimePtrCondition
			noTimeCond  NoTimeCondition
		)
		NowAndBeforeEach(func() {
			now = time.Unix(100, 0)
			metaNow = metav1.NewTime(now)
			acc = condition.NewAccessor(condition.AccessorOptions{
				Clock: clock.NewFakeClock(now),
			})
			deployCond = appsv1.DeploymentCondition{
				Type:               appsv1.DeploymentAvailable,
				Status:             corev1.ConditionTrue,
				LastUpdateTime:     metav1.Unix(2, 0),
				LastTransitionTime: metav1.Unix(1, 0),
				Reason:             "MinimumReplicasAvailable",
				Message:            "ReplicaSet \"foo\" has successfully progressed.",
			}
			deployConds = []appsv1.DeploymentCondition{deployCond}
			timePtrCondLastUpdateTime := metav1.Unix(2, 0)
			timePtrCondLastTransitionTime := metav1.Unix(1, 0)
			custCond = TimePtrCondition{
				Type:               "TimePtr",
				Status:             corev1.ConditionTrue,
				LastUpdateTime:     &timePtrCondLastUpdateTime,
				LastTransitionTime: &timePtrCondLastTransitionTime,
				Reason:             "MinimumReplicasAvailable",
				Message:            "ReplicaSet \"foo\" has successfully progressed.",
			}
			noTimeCond = NoTimeCondition{
				Type:    "NoTime",
				Status:  corev1.ConditionTrue,
				Reason:  "MinimumReplicasAvailable",
				Message: "ReplicaSet \"foo\" has successfully progressed.",
			}
		})

		Describe("Type", func() {
			It("should extract the type from a deployment condition (typed string)", func() {
				Expect(acc.Type(deployCond)).To(Equal(string(deployCond.Type)))
			})

			It("should extract the type from a custom condition (regular string)", func() {
				Expect(acc.Type(custCond)).To(Equal(custCond.Type))
			})
		})

		Describe("SetType", func() {
			It("should set the type on a deployment condition (typed string)", func() {
				Expect(acc.SetType(&deployCond, string(appsv1.DeploymentProgressing))).To(Succeed())
				Expect(deployCond.Type).To(Equal(appsv1.DeploymentProgressing))
			})

			It("should set the type on a custom condition (regular string)", func() {
				Expect(acc.SetType(&custCond, "OtherType")).To(Succeed())
				Expect(custCond.Type).To(Equal("OtherType"))
			})
		})

		Describe("Status", func() {
			It("should extract the status from a deployment condition (typed string)", func() {
				Expect(acc.Status(deployCond)).To(Equal(deployCond.Status))
			})

			It("should extract the status from a custom condition (regular string)", func() {
				Expect(acc.Status(custCond)).To(Equal(custCond.Status))
			})
		})

		Describe("SetStatus", func() {
			It("should set the status on a deployment condition (typed string)", func() {
				Expect(acc.SetStatus(&deployCond, corev1.ConditionFalse)).To(Succeed())
				Expect(deployCond.Status).To(Equal(corev1.ConditionFalse))
			})

			It("should set the status on a custom condition", func() {
				Expect(acc.SetStatus(&custCond, corev1.ConditionFalse)).To(Succeed())
				Expect(custCond.Status).To(Equal(corev1.ConditionFalse))
			})
		})

		Describe("LastUpdateTime", func() {
			It("should extract the lastUpdateTime from a deployment condition (typed string)", func() {
				Expect(acc.LastUpdateTime(deployCond)).To(Equal(deployCond.LastUpdateTime))
			})

			It("should extract the lastUpdateTime from a custom condition (regular string)", func() {
				Expect(acc.LastUpdateTime(custCond)).To(Equal(*custCond.LastUpdateTime))
			})
		})

		Describe("SetLastUpdateTime", func() {
			It("should set the lastUpdateTime on a deployment condition (typed string)", func() {
				Expect(acc.SetLastUpdateTime(&deployCond, metav1.Unix(10, 0))).To(Succeed())
				Expect(deployCond.LastUpdateTime).To(Equal(metav1.Unix(10, 0)))
			})

			It("should set the lastUpdateTime on a custom condition", func() {
				Expect(acc.SetLastUpdateTime(&custCond, metav1.Unix(10, 0))).To(Succeed())
				Expect(custCond.LastUpdateTime).To(PointTo(Equal(metav1.Unix(10, 0))))
			})
		})

		Describe("HasLastUpdateTime", func() {
			It("should return whether the given condition has a last update time field", func() {
				Expect(acc.HasLastUpdateTime(custCond)).To(BeTrue())
				Expect(acc.HasLastUpdateTime(noTimeCond)).To(BeFalse())
			})
		})

		Describe("SetLastUpdateTimeIfExists", func() {
			It("should set the last update time when the field exists", func() {
				Expect(acc.SetLastUpdateTimeIfExists(&custCond, metav1.Unix(10, 0))).To(Succeed())
				Expect(custCond.LastUpdateTime).To(PointTo(Equal(metav1.Unix(10, 0))))
			})

			It("should do nothing if the field does not exist", func() {
				Expect(acc.SetLastUpdateTimeIfExists(&noTimeCond, metav1.Unix(10, 0))).To(Succeed())
			})
		})

		Describe("LastTransitionTime", func() {
			It("should extract the lastTransitionTime from a deployment condition (typed string)", func() {
				Expect(acc.LastTransitionTime(deployCond)).To(Equal(deployCond.LastTransitionTime))
			})

			It("should extract the lastTransitionTime from a custom condition (regular string)", func() {
				Expect(acc.LastTransitionTime(custCond)).To(Equal(*custCond.LastTransitionTime))
			})
		})

		Describe("SetLastTransitionTime", func() {
			It("should set the lastTransitionTime on a deployment condition (typed string)", func() {
				Expect(acc.SetLastTransitionTime(&deployCond, metav1.Unix(10, 0))).To(Succeed())
				Expect(deployCond.LastTransitionTime).To(Equal(metav1.Unix(10, 0)))
			})

			It("should set the lastTransitionTime on a custom condition", func() {
				Expect(acc.SetLastTransitionTime(&custCond, metav1.Unix(10, 0))).To(Succeed())
				Expect(custCond.LastTransitionTime).To(PointTo(Equal(metav1.Unix(10, 0))))
			})
		})

		Describe("SetLastTransitionTimeIfExists", func() {
			It("should set the last transition time when the field exists", func() {
				Expect(acc.SetLastTransitionTimeIfExists(&custCond, metav1.Unix(10, 0))).To(Succeed())
				Expect(custCond.LastTransitionTime).To(PointTo(Equal(metav1.Unix(10, 0))))
			})

			It("should do nothing if the field does not exist", func() {
				Expect(acc.SetLastTransitionTimeIfExists(&noTimeCond, metav1.Unix(10, 0))).To(Succeed())
			})
		})

		Describe("Reason", func() {
			It("should extract the reason from a deployment condition (typed string)", func() {
				Expect(acc.Reason(deployCond)).To(Equal(deployCond.Reason))
			})

			It("should extract the reason from a custom condition (regular string)", func() {
				Expect(acc.Reason(custCond)).To(Equal(custCond.Reason))
			})
		})

		Describe("SetReason", func() {
			It("should set the reason on a deployment condition (typed string)", func() {
				Expect(acc.SetReason(&deployCond, "OtherReason")).To(Succeed())
				Expect(deployCond.Reason).To(Equal("OtherReason"))
			})

			It("should set the reason on a custom condition", func() {
				Expect(acc.SetReason(&custCond, "OtherReason")).To(Succeed())
				Expect(custCond.Reason).To(Equal("OtherReason"))
			})
		})

		Describe("Message", func() {
			It("should extract the message from a deployment condition (typed string)", func() {
				Expect(acc.Message(deployCond)).To(Equal(deployCond.Message))
			})

			It("should extract the message from a custom condition (regular string)", func() {
				Expect(acc.Message(custCond)).To(Equal(custCond.Message))
			})
		})

		Describe("SetMessage", func() {
			It("should set the message on a deployment condition (typed string)", func() {
				Expect(acc.SetMessage(&deployCond, "OtherMessage")).To(Succeed())
				Expect(deployCond.Message).To(Equal("OtherMessage"))
			})

			It("should set the message on a custom condition", func() {
				Expect(acc.SetMessage(&custCond, "OtherMessage")).To(Succeed())
				Expect(custCond.Message).To(Equal("OtherMessage"))
			})
		})

		Describe("Update", func() {
			It("should apply all updates and change the transition time in case the status changed", func() {
				Expect(acc.Update(&deployCond,
					condition.UpdateStatus(corev1.ConditionFalse),
					condition.UpdateReason("BadDay"),
					condition.UpdateMessage("Some message"),
				)).To(Succeed())
				Expect(deployCond).To(Equal(appsv1.DeploymentCondition{
					Type:               appsv1.DeploymentAvailable,
					Status:             corev1.ConditionFalse,
					LastUpdateTime:     metaNow,
					LastTransitionTime: metaNow,
					Reason:             "BadDay",
					Message:            "Some message",
				}))
			})

			It("should apply all updates and not change the transition time when the status did not change", func() {
				Expect(acc.Update(&deployCond,
					condition.UpdateReason("BadDay"),
					condition.UpdateMessage("Some message"),
				)).To(Succeed())
				Expect(deployCond).To(Equal(appsv1.DeploymentCondition{
					Type:               appsv1.DeploymentAvailable,
					Status:             corev1.ConditionTrue,
					LastUpdateTime:     metaNow,
					LastTransitionTime: metav1.Unix(1, 0),
					Reason:             "BadDay",
					Message:            "Some message",
				}))
			})
		})

		Describe("FindSlice", func() {
			It("should find the target condition if it's present", func() {
				var cond appsv1.DeploymentCondition
				Expect(acc.FindSlice([]appsv1.DeploymentCondition{deployCond}, string(appsv1.DeploymentAvailable), &cond)).To(BeTrue())
				Expect(cond).To(Equal(deployCond))
			})

			It("should not find the target condition if it's not present", func() {
				var cond appsv1.DeploymentCondition
				Expect(acc.FindSlice([]appsv1.DeploymentCondition{}, string(appsv1.DeploymentAvailable), &cond)).To(BeFalse())
				Expect(cond).NotTo(Equal(deployCond))
			})
		})

		Describe("FindSliceStatus", func() {
			It("should find the status of the target condition if it's present", func() {
				Expect(acc.FindSliceStatus(deployConds, string(appsv1.DeploymentAvailable))).To(Equal(corev1.ConditionTrue))
			})

			It("should return corev1.ConditionUnknown if the target condition is not present", func() {
				Expect(acc.FindSliceStatus(deployConds, string(appsv1.DeploymentProgressing))).To(Equal(corev1.ConditionUnknown))
			})
		})

		Describe("UpdateSlice", func() {
			It("should update the slice, adding the desired condition", func() {
				Expect(acc.UpdateSlice(&deployConds, string(appsv1.DeploymentProgressing),
					condition.UpdateStatus(corev1.ConditionFalse),
				)).To(Succeed())

				Expect(deployConds).To(Equal([]appsv1.DeploymentCondition{
					deployCond,
					{
						Type:               appsv1.DeploymentProgressing,
						Status:             corev1.ConditionFalse,
						LastUpdateTime:     metaNow,
						LastTransitionTime: metaNow,
					},
				}))
			})

			It("should update the slice, updating the existing condition in-place", func() {
				Expect(acc.UpdateSlice(&deployConds, string(appsv1.DeploymentAvailable),
					condition.UpdateStatus(corev1.ConditionFalse),
				)).To(Succeed())

				Expect(deployConds).To(Equal([]appsv1.DeploymentCondition{
					{
						Type:               appsv1.DeploymentAvailable,
						Status:             corev1.ConditionFalse,
						LastUpdateTime:     metaNow,
						LastTransitionTime: metaNow,
						Message:            deployCond.Message,
						Reason:             deployCond.Reason,
					},
				}))
			})
		})
	})
})
