package v7action_test

import (
	"errors"

	. "code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/actor/v7action/v7actionfakes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Buildpack", func() {
	var (
		actor                     *Actor
		fakeCloudControllerClient *v7actionfakes.FakeCloudControllerClient
	)

	BeforeEach(func() {
		actor, fakeCloudControllerClient, _, _, _ = NewTestActor()
	})

	Describe("DeleteBuildpack", func() {
		var (
			jobURL        JobURL
			buildpackGUID string
			warnings      Warnings
			executeErr    error
		)

		JustBeforeEach(func() {
			jobURL, warnings, executeErr = actor.DeleteBuildpack(buildpackGUID)
		})

		When("deleting a buildpack fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.DeleteBuildpackReturns(
					"",
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("some-error"))
			})

			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("some-error"))
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
				Expect(fakeCloudControllerClient.DeleteBuildpackCallCount()).To(Equal(1))
				paramGUID := fakeCloudControllerClient.DeleteBuildpackArgsForCall(0)
				Expect(paramGUID).To(Equal(buildpackGUID))
			})
		})

		When("deleting the buildpack is successful", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.DeleteBuildpackReturns(
					JobURL("some-job-url"),
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					nil)
			})

			It("returns the jobURL and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
				Expect(jobURL).To(Equal(JobURL("some-job-url")))
				Expect(fakeCloudControllerClient.DeleteBuildpackCallCount()).To(Equal(1))
				paramGUID := fakeCloudControllerClient.DeleteBuildpackArgsForCall(0)
				Expect(paramGUID).To(Equal(buildpackGUID))
			})
		})
	})

	Describe("GetBuildpackByNameAndStack", func() {
		var (
			buildpackName  string
			buildpackStack string
			buildpack      Buildpack
			warnings       Warnings
			executeErr     error
		)

		JustBeforeEach(func() {
			buildpack, warnings, executeErr = actor.GetBuildpackByNameAndStack(buildpackName, buildpackStack)
		})

		When("getting buildpacks fails", func() {
			BeforeEach(func() {

				buildpackStack = "real-good-stack"
				fakeCloudControllerClient.GetBuildpacksReturns(
					nil,
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("some-error"))
			})

			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("some-error"))
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
				Expect(fakeCloudControllerClient.GetBuildpacksCallCount()).To(Equal(1))
				queries := fakeCloudControllerClient.GetBuildpacksArgsForCall(0)
				Expect(queries).To(ConsistOf(
					ccv3.Query{
						Key:    ccv3.NameFilter,
						Values: []string{buildpackName},
					},
					ccv3.Query{
						Key:    ccv3.StackFilter,
						Values: []string{buildpackStack},
					},
				))
			})
		})

		When("getting buildpacks is successful", func() {
			When("No stack is specified", func() {
				BeforeEach(func() {
					buildpackStack = ""
					buildpackName = "my-buildpack"

					ccBuildpack := ccv3.Buildpack{Name: "my-buildpack", GUID: "some-guid"}
					fakeCloudControllerClient.GetBuildpacksReturns(
						[]ccv3.Buildpack{ccBuildpack},
						ccv3.Warnings{"some-warning-1", "some-warning-2"},
						nil)
				})

				It("Returns the proper buildpack", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
					Expect(buildpack).To(Equal(Buildpack{Name: "my-buildpack", GUID: "some-guid"}))
				})

				It("Does not pass a stack query to the client", func() {
					Expect(fakeCloudControllerClient.GetBuildpacksCallCount()).To(Equal(1))
					queries := fakeCloudControllerClient.GetBuildpacksArgsForCall(0)
					Expect(queries).To(ConsistOf(
						ccv3.Query{
							Key:    ccv3.NameFilter,
							Values: []string{buildpackName},
						},
					))
				})
			})

			When("A stack is specified", func() {
				BeforeEach(func() {
					buildpackStack = "good-stack"
					buildpackName = "my-buildpack"

					ccBuildpack := ccv3.Buildpack{Name: "my-buildpack", GUID: "some-guid", Stack: "good-stack"}
					fakeCloudControllerClient.GetBuildpacksReturns(
						[]ccv3.Buildpack{ccBuildpack},
						ccv3.Warnings{"some-warning-1", "some-warning-2"},
						nil)
				})

				It("Returns the proper buildpack", func() {
					Expect(executeErr).ToNot(HaveOccurred())
					Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
					Expect(buildpack).To(Equal(Buildpack{Name: "my-buildpack", GUID: "some-guid", Stack: "good-stack"}))
				})

				It("Does pass a stack query to the client", func() {
					Expect(fakeCloudControllerClient.GetBuildpacksCallCount()).To(Equal(1))
					queries := fakeCloudControllerClient.GetBuildpacksArgsForCall(0)
					Expect(queries).To(ConsistOf(
						ccv3.Query{
							Key:    ccv3.NameFilter,
							Values: []string{buildpackName},
						},
						ccv3.Query{
							Key:    ccv3.StackFilter,
							Values: []string{buildpackStack},
						},
					))
				})
			})
		})
	})

	Describe("GetBuildpacks", func() {
		var (
			buildpacks []Buildpack
			warnings   Warnings
			executeErr error
		)

		JustBeforeEach(func() {
			buildpacks, warnings, executeErr = actor.GetBuildpacks()
		})

		When("getting buildpacks fails", func() {
			BeforeEach(func() {
				fakeCloudControllerClient.GetBuildpacksReturns(
					nil,
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					errors.New("some-error"))
			})

			It("returns warnings and error", func() {
				Expect(executeErr).To(MatchError("some-error"))
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
			})
		})

		When("getting buildpacks is successful", func() {
			BeforeEach(func() {
				ccBuildpacks := []ccv3.Buildpack{
					{Name: "buildpack-1", Position: 1},
					{Name: "buildpack-2", Position: 2},
				}

				fakeCloudControllerClient.GetBuildpacksReturns(
					ccBuildpacks,
					ccv3.Warnings{"some-warning-1", "some-warning-2"},
					nil)
			})

			It("returns the buildpacks and warnings", func() {
				Expect(executeErr).ToNot(HaveOccurred())
				Expect(warnings).To(ConsistOf("some-warning-1", "some-warning-2"))
				Expect(buildpacks).To(Equal([]Buildpack{
					{Name: "buildpack-1", Position: 1},
					{Name: "buildpack-2", Position: 2},
				}))

				Expect(fakeCloudControllerClient.GetBuildpacksCallCount()).To(Equal(1))
				Expect(fakeCloudControllerClient.GetBuildpacksArgsForCall(0)).To(ConsistOf(ccv3.Query{
					Key:    ccv3.OrderBy,
					Values: []string{ccv3.PositionOrder},
				}))
			})
		})
	})

})
