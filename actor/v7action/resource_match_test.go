package v7action_test

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/cf/errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Resource Matching", func() {
	var (
		resources                 []sharedaction.V3Resource
		executeErr                error
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
		actor                     *Actor
		fakeSharedActor           *v7actionfakes.FakeSharedActor
		fakeConfig                *v7actionfakes.FakeConfig

		matchedResources []sharedaction.V3Resource

		warnings Warnings
	)

	BeforeEach(func() {
		fakeCloudControllerClient = new(v7actionfakes.FakeCloudControllerClient)
		fakeConfig = new(v7actionfakes.FakeConfig)
		fakeSharedActor = new(v7actionfakes.FakeSharedActor)
		actor = NewActor(fakeCloudControllerClient, fakeConfig, fakeSharedActor, nil)
	})

	JustBeforeEach(func() {
		matchedResources, warnings, executeErr = actor.ResourceMatch(resources)
	})

	When("The cc client succeeds", func() {
		BeforeEach(func() {
			resources = []sharedaction.V3Resource{
				{FilePath: "path/to/file"},
				{FilePath: "path/to/file2"},
			}

			fakeCloudControllerClient.ResourceMatchReturns([]ccv3.Resource{{FilePath: "path/to/file"}}, ccv3.Warnings{"this-is-a-warning"}, nil)
		})
		It("passes through the list of resources", func() {
			Expect(fakeCloudControllerClient.ResourceMatchCallCount()).To(Equal(1))
			Expect(executeErr).ToNot(HaveOccurred())
			Expect(warnings).To(ConsistOf(ccv3.Warnings{"this-is-a-warning"}))
			passedResources := fakeCloudControllerClient.ResourceMatchArgsForCall(0)
			Expect(passedResources).To(ConsistOf(
				ccv3.Resource{FilePath: "path/to/file"},
				ccv3.Resource{FilePath: "path/to/file2"},
			))
		})

		It("returns a list of sharedAction V3Resources", func() {
			Expect(matchedResources).To(ConsistOf(sharedaction.V3Resource{FilePath: "path/to/file"}))
		})
	})

	When("The cc client errors", func() {
		BeforeEach(func() {
			fakeCloudControllerClient.ResourceMatchReturns([]ccv3.Resource{}, ccv3.Warnings{"this-is-a-warning"}, errors.New("boom"))
		})

		It("raises the error", func() {
			Expect(executeErr).To(MatchError(errors.New("boom")))
			Expect(warnings).To(ConsistOf(ccv3.Warnings{"this-is-a-warning"}))
		})
	})
})
