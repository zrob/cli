package v7

import (
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/command/flag"
)

//go:generate counterfeiter . DeleteBuildpackActor

type DeleteBuildpackActor interface {
	DeleteBuildpackByNameAndStack(buildpackName string, buildpackStack string) (v7action.Warnings, error)
}

type DeleteBuildpackCommand struct {
	RequiredArgs flag.BuildpackName `positional-args:"yes"`
	Force        bool               `short:"f" description:"Force deletion without confirmation"`
	Stack        string             `short:"s" description:"Specify stack to disambiguate buildpacks with the same name. Required when buildpack name is ambiguous"`
	Actor        DeleteBuildpackActor
	UI           command.UI
	Config       command.Config
	SharedActor  command.SharedActor
}

func (cmd *DeleteBuildpackCommand) Setup(config command.Config, ui command.UI) error {
	return nil
}

func (cmd DeleteBuildpackCommand) Execute(args []string) error {
	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	if !cmd.Force {
		response, err := cmd.UI.DisplayBoolPrompt(false, "Really delete the {{.ModelType}} {{.ModelName}}?", map[string]interface{}{
			"ModelType": "buildpack",
			"ModelName": cmd.RequiredArgs.Buildpack,
		})
		if err != nil {
			return err
		}

		if !response {
			cmd.UI.DisplayText("Delete cancelled")
			return nil
		}
	}

	if cmd.Stack == "" {
		cmd.UI.DisplayTextWithFlavor("Deleting buildpack {{.BuildpackName}}...", map[string]interface{}{
			"BuildpackName": cmd.RequiredArgs.Buildpack,
		})

	} else {
		cmd.UI.DisplayTextWithFlavor("Deleting buildpack {{.BuildpackName}} with stack {{.Stack}}...", map[string]interface{}{
			"BuildpackName": cmd.RequiredArgs.Buildpack,
			"Stack":         cmd.Stack,
		})
	}
	warnings, err := cmd.Actor.DeleteBuildpackByNameAndStack(cmd.RequiredArgs.Buildpack, cmd.Stack)
	cmd.UI.DisplayWarnings(warnings)

	//Fetch:
	// actionerror.BuildpackNotFoundError{}
	// actionerror.MultipleBuildpacksFoundError{}


	if err != nil {
		return err
	}
	cmd.UI.DisplayOK()

	return nil
}
