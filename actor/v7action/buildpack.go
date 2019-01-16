package v7action

import "code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"

type Buildpack ccv3.Buildpack
type JobURL ccv3.JobURL

func (actor Actor) GetBuildpacks() ([]Buildpack, Warnings, error) {
	ccv3Buildpacks, warnings, err := actor.CloudControllerClient.GetBuildpacks(ccv3.Query{
		Key:    ccv3.OrderBy,
		Values: []string{ccv3.PositionOrder},
	})

	var buildpacks []Buildpack
	for _, buildpack := range ccv3Buildpacks {
		buildpacks = append(buildpacks, Buildpack(buildpack))
	}

	return buildpacks, Warnings(warnings), err
}

func (actor Actor) GetBuildpackByNameAndStack(buildpackName string, buildpackStack string) (Buildpack, Warnings, error) {
	var (
		ccv3Buildpacks []ccv3.Buildpack
		warnings       ccv3.Warnings
		err            error
	)

	if buildpackStack == "" {
		ccv3Buildpacks, warnings, err = actor.CloudControllerClient.GetBuildpacks(ccv3.Query{
			Key:    ccv3.NameFilter,
			Values: []string{buildpackName},
		})
	} else {
		ccv3Buildpacks, warnings, err = actor.CloudControllerClient.GetBuildpacks(
			ccv3.Query{
				Key:    ccv3.NameFilter,
				Values: []string{buildpackName},
			},
			ccv3.Query{
				Key:    ccv3.StackFilter,
				Values: []string{buildpackStack},
			},
		)
	}

	if err != nil {
		return Buildpack{}, Warnings(warnings), err
	}

	if len(ccv3Buildpacks) == 0 {
		return Buildpack{}, Warnings(warnings), err
	}

	return Buildpack(ccv3Buildpacks[0]), Warnings(warnings), err
}

func (actor Actor) DeleteBuildpack(buildpackGUID string) (JobURL, Warnings, error) {
	jobURL, warnings, err := actor.CloudControllerClient.DeleteBuildpack(buildpackGUID)

	return jobURL, Warnings(warnings), err
}
