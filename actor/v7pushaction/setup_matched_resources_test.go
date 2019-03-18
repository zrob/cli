package v7pushaction_test

import (
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("SetupAllResourcesForPushPlan", func() {
	var (
		actor           *Actor
		fakeSharedActor *v7pushactionfakes.FakeSharedActor
		fakeV7Actor     *v7pushactionfakes.FakeV7Actor

		pushPlan PushPlan

		expectedPushPlan PushPlan
		executeErr       error
	)

	BeforeEach(func() {
		actor, _, fakeV7Actor, fakeSharedActor = getTestPushActor()

		pushPlan = PushPlan{}
	})

	JustBeforeEach(func() {
		expectedPushPlan, executeErr = actor.SetupMatchedResourcesForPushPlan(pushPlan)
	})

	When("the application is a docker app", func() {
		BeforeEach(func() {
			pushPlan.Application.LifecycleType = constant.AppLifecycleTypeDocker
		})

		It("skips matching the resources", func() {
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(pushPlan.MatchedResources).To(BeEmpty())
			Expect(pushPlan.UnmatchedResources).To(BeEmpty())

			Expect(fakeSharedActor.MatchResourcesCallCount()).To(Equal(0))
		})
	})

	When("the application is a buildpack app", func() {
		When("push plan has no resources is not set", func() {
			It("returns an error", func() {
			})
		})

		When("there are resources", func() {
			BeforeEach(func() {
				pushPlan = PushPlan{
					AllResources: []sharedaction.V3Resource{
						{FilePath: "invader zim"},
						{FilePath: "uncle tito"},
					},
				}

				fakeV7Actor.MatchedResourcesReturns([]sharedaction.V3Resource{
					{FilePath: "uncle tito"},
				})
			})

			It("Resource matches", func() {
				Expect(fakeV7Actor.ResourceMatchCallCount()).To(Equal(1))
				passedResources := fakeV7Actor.ResourceMatchArgsForCall(0)
				Expect(passedResources).To(ConsistOf(
					sharedaction.V3Resource{FilePath: "invader zim"},
					sharedaction.V3Resource{FilePath: "uncle tito"},
				))
			})

			It("sets the matched resources and unmatched resources", func() {
				Expect(expectedPushPlan.MatchedResources).To(ConsistOf(
					sharedaction.V3Resource{FilePath: "uncle tito"},
				))

				Expect(expectedPushPlan.UnmatchedResources).To(ConsistOf(
					sharedaction.V3Resource{FilePath: "invader zim"},
				))
			})

		})

		When("matching resources errors", func() {

		})
	})

})
