package v7

import (
	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
	"code.cloudfoundry.org/cli/command/translatableerror"
	"code.cloudfoundry.org/cli/command/v7/shared"
	"code.cloudfoundry.org/cli/types"
	"code.cloudfoundry.org/cli/util/download"
	"io/ioutil"
	"os"
	"time"
)

//go:generate counterfeiter . UpdateBuildpackActor

type UpdateBuildpackActor interface {
	UpdateBuildpackByNameAndStack(buildpackName string, buildpackStack string, buildpack v7action.Buildpack) (v7action.Buildpack, v7action.Warnings, error)
	UploadBuildpack(guid string, pathToBuildpackBits string, progressBar v7action.SimpleProgressBar) (ccv3.JobURL, v7action.Warnings, error)
	PrepareBuildpackBits(inputPath string, tmpDirPath string, downloader v7action.Downloader) (string, error)
	PollUploadBuildpackJob(jobURL ccv3.JobURL) (v7action.Warnings, error)
}

type UpdateBuildpackCommand struct {
	RequiredArgs    flag.UpdateBuildpackArgs         `positional-args:"Yes"`
	usage           interface{}                      `usage:"CF_NAME update-buildpack BUILDPACK [-p PATH | -s STACK | --assign-stack NEW_STACK] [-i POSITION] [--enable|--disable] [--lock|--unlock]\n\nTIP:\nPath should be a zip file, a url to a zip file, or a local directory. Position is a positive integer, sets priority, and is sorted from lowest to highest.\n\nUse '--assign-stack' with caution. Associating a buildpack with a stack that it does not support may result in undefined behavior. Additionally, changing this association once made may require a local copy of the buildpack.\n\n"`
	relatedCommands interface{}                      `related_commands:"buildpacks, create-buildpack, delete-buildpack"`
	NewStack        string                           `long:"assign-stack" description:"Assign a stack to a buildpack that does not have a stack association"`
	Disable         bool                             `long:"disable" description:"Disable the buildpack from being used for staging"`
	Enable          bool                             `long:"enable" description:"Enable the buildpack to be used for staging"`
	Lock            bool                             `long:"lock" description:"Lock the buildpack to prevent updates"`
	Path            flag.PathWithExistenceCheckOrURL `long:"path" short:"p" description:"Path to directory or zip file"`
	Position        types.NullInt                    `long:"position" short:"i" description:"The order in which the buildpacks are checked during buildpack auto-detection"`
	CurrentStack    string                           `long:"stack" short:"s" description:"Specify stack to disambiguate buildpacks with the same name"`
	Unlock          bool                             `long:"unlock" description:"Unlock the buildpack to enable updates"`

	UI          command.UI
	Config      command.Config
	ProgressBar v7action.SimpleProgressBar
	SharedActor command.SharedActor
	Actor       UpdateBuildpackActor
}

func (cmd *UpdateBuildpackCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	sharedActor := sharedaction.NewActor(config)
	cmd.SharedActor = sharedActor

	ccClient, uaaClient, err := shared.NewClients(config, ui, true, "")
	if err != nil {
		return err
	}
	cmd.Actor = v7action.NewActor(ccClient, config, sharedActor, uaaClient)
	cmd.ProgressBar = v7action.NewProgressBar()

	return nil
}

func (cmd UpdateBuildpackCommand) Execute(args []string) error {
	var buildpackBitsPath string

	err := cmd.validateFlagCombinations()
	if err != nil {
		return err
	}

	err = cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.printInitialText(user.Name)

	if cmd.Path != "" {
		var (
			tmpDirPath string
		)
		downloader := download.NewDownloader(time.Second * 30)
		tmpDirPath, err = ioutil.TempDir("", "buildpack-dir-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(tmpDirPath)

		buildpackBitsPath, err = cmd.Actor.PrepareBuildpackBits(string(cmd.Path), tmpDirPath, downloader)
		if err != nil {
			return err
		}
	}

	var desiredBuildpack v7action.Buildpack

	desiredBuildpack.Enabled = types.NullBool{IsSet: cmd.Enable || cmd.Disable, Value: cmd.Enable}
	desiredBuildpack.Locked = types.NullBool{IsSet: cmd.Lock || cmd.Unlock, Value: cmd.Lock}
	desiredBuildpack.Position = cmd.Position

	if cmd.NewStack != "" {
		desiredBuildpack.Stack = cmd.NewStack
	}

	updatedBuildpack, warnings, err := cmd.Actor.UpdateBuildpackByNameAndStack(
		cmd.RequiredArgs.Buildpack,
		cmd.CurrentStack,
		desiredBuildpack,
	)
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	if buildpackBitsPath != "" {
		cmd.UI.DisplayTextWithFlavor("Uploading buildpack {{.Buildpack}} as {{.Username}}...", map[string]interface{}{
			"Buildpack": cmd.RequiredArgs.Buildpack,
			"Username":  user.Name,
		})

		jobURL, warnings, err := cmd.Actor.UploadBuildpack(
			updatedBuildpack.GUID,
			buildpackBitsPath,
			cmd.ProgressBar,
		)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}
		cmd.UI.DisplayOK()

		cmd.UI.DisplayTextWithFlavor("Processing uploaded buildpack {{.BuildpackName}}...", map[string]interface{}{
			"BuildpackName": cmd.RequiredArgs.Buildpack,
		})

		warnings, err = cmd.Actor.PollUploadBuildpackJob(jobURL)
		cmd.UI.DisplayWarnings(warnings)
		if err != nil {
			return err
		}

		cmd.UI.DisplayOK()
	}
	return err
}

func (cmd UpdateBuildpackCommand) printInitialText(userName string) {
	if cmd.NewStack != "" {
		cmd.UI.DisplayTextWithFlavor("Assigning stack {{.Stack}} to {{.Buildpack}} as {{.CurrentUser}}...", map[string]interface{}{
			"Buildpack":   cmd.RequiredArgs.Buildpack,
			"CurrentUser": userName,
			"Stack":       cmd.NewStack,
		})
		if cmd.Position.IsSet || cmd.Lock || cmd.Unlock || cmd.Enable || cmd.Disable {
			cmd.UI.DisplayTextWithFlavor("Updating buildpack {{.Buildpack}} with stack {{.Stack}} as {{.CurrentUser}}...", map[string]interface{}{
				"Buildpack":   cmd.RequiredArgs.Buildpack,
				"CurrentUser": userName,
				"Stack":       cmd.NewStack,
			})
		}
	} else if cmd.CurrentStack == "" {
		cmd.UI.DisplayTextWithFlavor("Updating buildpack {{.Buildpack}} as {{.CurrentUser}}...", map[string]interface{}{
			"Buildpack":   cmd.RequiredArgs.Buildpack,
			"CurrentUser": userName,
		})
	} else {
		cmd.UI.DisplayTextWithFlavor("Updating buildpack {{.Buildpack}} with stack {{.Stack}} as {{.CurrentUser}}...", map[string]interface{}{
			"Buildpack":   cmd.RequiredArgs.Buildpack,
			"CurrentUser": userName,
			"Stack":       cmd.CurrentStack,
		})
	}
}

func (cmd UpdateBuildpackCommand) validateFlagCombinations() error {
	if cmd.Lock && cmd.Unlock {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--lock", "--unlock"},
		}
	}

	if cmd.Enable && cmd.Disable {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--enable", "--disable"},
		}
	}

	if cmd.Path != "" && cmd.Lock {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--path", "--lock"},
		}
	}

	if cmd.Path != "" && cmd.NewStack != "" {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--path", "--assign-stack"},
		}
	}

	if cmd.CurrentStack != "" && cmd.NewStack != "" {
		return translatableerror.ArgumentCombinationError{
			Args: []string{"--stack", "--assign-stack"},
		}
	}
	return nil
}
