package v7action

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
)

func (actor Actor) ResourceMatch(resources []sharedaction.V3Resource) ([]sharedaction.V3Resource, Warnings, error) {
	resourceChunks := actor.chunkResources(resources)

	var (
		allWarnings         Warnings
		matchedApiResources []ccv3.Resource
	)
	for _, chunk := range resourceChunks {
		newMatchedApiResources, warnings, err := actor.CloudControllerClient.ResourceMatch(chunk)
		allWarnings = append(allWarnings, warnings...)
		if err != nil {
			return nil, allWarnings, err
		}
		matchedApiResources = append(matchedApiResources, newMatchedApiResources...)
	}

	var matchedResources []sharedaction.V3Resource
	for _, resource := range matchedApiResources {
		matchedResources = append(matchedResources, sharedaction.V3Resource(resource))
	}

	return matchedResources, allWarnings, nil
}

func (Actor) chunkResources(resources []sharedaction.V3Resource) [][]ccv3.Resource {
	var chunkedResources [][]ccv3.Resource
	var currentSet []ccv3.Resource

	for index, resource := range resources {
		currentSet = append(currentSet, ccv3.Resource(resource))
		if len(currentSet) == constant.MaxNumberOfResourcesForMatching || index+1 == len(resources) {
			chunkedResources = append(chunkedResources, currentSet)
			currentSet = []ccv3.Resource{}
		}
	}
	return chunkedResources
}
