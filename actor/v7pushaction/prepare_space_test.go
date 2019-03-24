package v7pushaction_test

import (
	"errors"

	"code.cloudfoundry.org/cli/actor/actionerror"
	"code.cloudfoundry.org/cli/actor/v7action"

	. "code.cloudfoundry.org/cli/actor/v7pushaction"
	"code.cloudfoundry.org/cli/actor/v7pushaction/v7pushactionfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func PrepareSpaceStreamsDrainedAndClosed(
	eventStream <-chan Event,
) bool {
	var configStreamClosed, eventStreamClosed, warningsStreamClosed, errorStreamClosed bool
	for {
		if _, ok := <-eventStream; !ok {
			return true
		}
	}
}

var _ = FDescribe("PrepareSpace", func() {
	var (
		actor       *Actor
		fakeV7Actor *v7pushactionfakes.FakeV7Actor

		pushPlans          []PushPlan
		fakeManifestParser *v7pushactionfakes.FakeManifestParser

		spaceGUID string

		eventStream <-chan Event
	)

	BeforeEach(func() {
		actor, _, fakeV7Actor, _ = getTestPushActor()

		spaceGUID = "space"

		fakeManifestParser = new(v7pushactionfakes.FakeManifestParser)
	})

	AfterEach(func() {
		Eventually(PrepareSpaceStreamsDrainedAndClosed(eventStream)).Should(BeTrue())
	})

	JustBeforeEach(func() {
		eventStream = actor.PrepareSpace(pushPlans, fakeManifestParser)
	})

	FWhen("there is a single push state and no manifest", func() {
		var appName = "app-name"
		BeforeEach(func() {
			fakeManifestParser.FullRawManifestReturns(nil)
		})
		When("Creating the app succeeds", func() {
			BeforeEach(func() {
				pushPlans = []PushPlan{{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName}}}
				fakeV7Actor.CreateApplicationInSpaceReturns(
					v7action.Application{Name: appName},
					v7action.Warnings{"create-app-warning"},
					nil,
				)
			})
			It("creates the app using the API", func() {
				Consistently(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(0))
				Eventually(eventStream).Should(Recieve(Equal(Event{Type: CreatingApplication})))
				Eventually(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(1))
				actualApp, actualSpaceGUID := fakeV7Actor.CreateApplicationInSpaceArgsForCall(0)
				Expect(actualApp).To(Equal(v7action.Application{Name: appName}))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Eventually(eventStream).Should(Receive(Equal(Event{
					Type:     CreatedApplication,
					Warnings: Warnings{"create-app-warning"},
					Data: []PushPlan{
						{
							SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName},
						},
					},
				})))
			})
		})

		When("the app already exists", func() {
			BeforeEach(func() {
				pushPlans = []PushPlan{{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName}}}
				fakeV7Actor.CreateApplicationInSpaceReturns(
					v7action.Application{},
					v7action.Warnings{"create-app-warning"},
					actionerror.ApplicationAlreadyExistsError{},
				)
			})
			It("Sends already exists events", func() {
				Consistently(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(0))
				Eventually(eventStream).Should(Recieve(Equal(Event{Type: SkippingApplicationCreation})))
				Eventually(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(1))
				actualApp, actualSpaceGUID := fakeV7Actor.CreateApplicationInSpaceArgsForCall(0)
				Expect(actualApp).To(Equal(v7action.Application{Name: appName}))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Eventually(eventStream).Should(Receive(Equal(Event{
					Type:     ApplicationAlreadyExists,
					Warnings: Warnings{"create-app-warning"},
					Data: []PushPlan{
						{
							SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName},
						},
					},
				})))
			})
		})

		When("creating the app fails", func() {
			BeforeEach(func() {
				pushPlans = []PushPlan{{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName}}}
				fakeV7Actor.CreateApplicationInSpaceReturns(
					v7action.Application{},
					v7action.Warnings{"create-app-warning"},
					errors.New("some-create-error"),
				)
			})

			It("Returns the error", func() {
				Consistently(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(0))
				Eventually(eventStream).Should(Recieve(Equal(Event{Type: CreatingApplication})))
				Eventually(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(1))
				actualApp, actualSpaceGuid := fakeV7Actor.CreateApplicationInSpaceArgsForCall(0)
				Expect(actualApp.Name).To(Equal(appName))
				Expect(actualSpaceGuid).To(Equal(spaceGUID))
				Eventually(eventStream).Should(Receive(Equal(Event{
					Type:     Error,
					Warnings: Warnings{"create-app-warning"},
					Data: []PushPlan{
						{
							SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName},
						},
					},
					Err: errors.New("some-create-error"),
				})))
			})
		})
	})

	When("There is a a manifest", func() {
		var (
			manifest = []byte("app manifest")
			appName1 = "app-name1"
			appName2 = "app-name2"
		)
		When("applying the manifest fails", func() {
			BeforeEach(func() {
				pushPlans = []PushPlan{{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName1}}}
				fakeManifestParser.FullRawManifestReturns(manifest)
				fakeManifestParser.RawAppManifestReturns(manifest, nil)
				fakeV7Actor.SetSpaceManifestReturns(v7action.Warnings{"apply-manifest-warnings"}, errors.New("some-error"))
			})

			It("returns the error and exits", func() {
				Consistently(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(0))
				Eventually(eventStream).Should(Recieve(Equal(Event{Type: ApplyManifest})))
				Eventually(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(1))
				actualSpaceGuid, actualManifest := fakeV7Actor.SetSpaceManifestArgsForCall(0)
				Expect(actualSpaceGuid).To(Equal(spaceGUID))
				Expect(actualManifest).To(Equal(manifest))
				Eventually(eventStream).Should(Receive(Equal(Event{
					Type:     Error,
					Warnings: Warnings{"apply-manifest-warnings"},
					Data: []PushPlan{
						{
							SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName},
						},
					},
					Err: errors.New("some-error"),
				})))
			})
		})

		When("There is a single pushPlan", func() {

			BeforeEach(func() {
				pushPlans = []PushPlan{{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName1}}}
				fakeManifestParser.FullRawManifestReturns(manifest)
				fakeManifestParser.RawAppManifestReturns(manifest, nil)
				fakeV7Actor.SetSpaceManifestReturns(v7action.Warnings{"apply-manifest-warnings"}, nil)
			})

			It("applies the app specific manifest", func() {
				Consistently(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(0))
				Consistently(fakeManifestParser.FullRawManifestCallCount).Should(Equal(1))
				Eventually(eventStream).Should(Recieve(Equal(Event{Type: ApplyManifest})))
				Eventually(fakeManifestParser.RawAppManifestCallCount).Should(Equal(1))
				actualAppName := fakeManifestParser.RawAppManifestArgsForCall(0)
				Expect(actualAppName).To(Equal(appName1))
				Eventually(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(1))
				actualSpaceGUID, actualManifest := fakeV7Actor.SetSpaceManifestArgsForCall(0)
				Expect(actualManifest).To(Equal(manifest))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Eventually(eventStream).Should(Receive(Equal(Event{
					Type:     ApplyManifestComplete,
					Warnings: Warnings{"apply-manifest-warnings"},
					Data: []PushPlan{
						{
							SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName1},
						},
					},
				})))
			})
		})

		When("There are multiple push states", func() {
			BeforeEach(func() {
				pushPlans = []PushPlan{
					{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName1}},
					{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName2}},
				}
				fakeManifestParser.FullRawManifestReturns(manifest)
				fakeV7Actor.SetSpaceManifestReturns(v7action.Warnings{"apply-manifest-warnings"}, nil)
			})

			It("Applies the entire manifest", func() {
				Consistently(fakeV7Actor.CreateApplicationInSpaceCallCount).Should(Equal(0))
				Consistently(fakeManifestParser.RawAppManifestCallCount).Should(Equal(0))
				Eventually(eventStream).Should(Recieve(Equal(Event{Type: ApplyManifest})))
				Eventually(fakeManifestParser.FullRawManifestCallCount).Should(Equal(2))
				Eventually(fakeV7Actor.SetSpaceManifestCallCount).Should(Equal(1))
				actualSpaceGUID, actualManifest := fakeV7Actor.SetSpaceManifestArgsForCall(0)
				Expect(actualManifest).To(Equal(manifest))
				Expect(actualSpaceGUID).To(Equal(spaceGUID))
				Eventually(eventStream).Should(Receive(Equal(Event{
					Type:     ApplyManifestComplete,
					Warnings: Warnings{"apply-manifest-warnings"},
					Data: []PushPlan{
						{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName1}},
						{SpaceGUID: spaceGUID, Application: v7action.Application{Name: appName2}},
					},
				})))
			})
		})
	})

})
