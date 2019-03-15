package v7action

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
)

func (actor *Actor) ResourceMatch(resources []sharedaction.V3Resource) ([]sharedaction.V3Resource, Warnings,  error) {
	var resourcesCC []ccv3.Resource
	for _, resource := range resources {
		resourcesCC = append(resourcesCC, ccv3.Resource(resource))
	}

	matchedApiResources, warnings, err := actor.CloudControllerClient.ResourceMatch(resourcesCC)

	var matchedResources []sharedaction.V3Resource

	for _, resource := range matchedApiResources {
		matchedResources = append(matchedResources, sharedaction.V3Resource(resource))
	}

	return matchedResources, Warnings(warnings), err
}
