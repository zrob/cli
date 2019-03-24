package v7pushaction

import (
	"code.cloudfoundry.org/cli/actor/actionerror"
	log "github.com/sirupsen/logrus"
)

func (actor Actor) PrepareSpace(pushPlans []PushPlan, manifestParser ManifestParser) <-chan Event {
	eventStream := make(chan Event)

	go func() {
		log.Debug("starting space preparation go routine")
		defer close(eventStream)

		if manifestParser.FullRawManifest() == nil {
			_, warnings, err := actor.V7Actor.CreateApplicationInSpace(pushPlans[0].Application, pushPlans[0].SpaceGUID)
			if _, ok := err.(actionerror.ApplicationAlreadyExistsError); ok {
				eventStream <- NewEvent(SkippingApplicationCreation, nil, nil, pushPlans)
				eventStream <- NewEvent(ApplicationAlreadyExists, warnings, nil, pushPlans)
			} else {
				eventStream <- NewEvent(CreatingApplication, nil, nil, pushPlans)
				eventStream <- NewEvent(CreatedApplication, warngings, err, pushPlans)
			}
		} else {
			var manifest []byte
			manifest, err := getManifest(pushPlans, manifestParser)
			if err != nil {
				eventStream <- NewEvent(Error, nil, err, pushPlans)
				return
			}
			eventStream <- NewEvent(ApplyManifest, nil, nil, pushPlans)
			warnings, err := actor.V7Actor.SetSpaceManifest(pushPlans[0].SpaceGUID, manifest) // CAN WE HAVE AN EMPTY MANIFEST
			eventStream <- NewEvent(ApplyManifestComplete, warnings, err, pushPlans)
		}

	}()

	return eventStream
}

func getManifest(plans []PushPlan, parser ManifestParser) ([]byte, error) {
	if len(plans) == 1 {
		return parser.RawAppManifest(plans[0].Application.Name)
	}
	return parser.FullRawManifest(), nil
}
