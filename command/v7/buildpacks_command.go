package v7

import (
	"strconv"

	"code.cloudfoundry.org/cli/actor/sharedaction"
	"code.cloudfoundry.org/cli/actor/v7action"
	"code.cloudfoundry.org/cli/command"
	"code.cloudfoundry.org/cli/util/ui"
)

//go:generate counterfeiter . BuildpacksActor

type BuildpacksActor interface {
	GetBuildpacks() ([]v7action.BuildpackTemp, v7action.Warnings, error)
}

type BuildpacksCommand struct {
	usage           interface{} `usage:"CF_NAME buildpacks"`
	relatedCommands interface{} `related_commands:"push"`

	UI          command.UI
	Config      command.Config
	SharedActor command.SharedActor
	Actor       BuildpacksActor
}

func (cmd *BuildpacksCommand) Setup(config command.Config, ui command.UI) error {
	cmd.UI = ui
	cmd.Config = config
	cmd.SharedActor = sharedaction.NewActor(config)

	// ccClient, _, err := shared.NewClients(config, ui, true, "")
	// if err != nil {
	// 	return err
	// }
	// cmd.Actor = v7action.NewActor(ccClient, config, nil, nil)

	return nil
}

func (cmd BuildpacksCommand) Execute(args []string) error {
	// const MaxArgsAllowed = 0
	// if len(args) > MaxArgsAllowed {
	// 	return translatableerror.TooManyArgumentsError{
	// 		ExtraArgument: args[MaxArgsAllowed],
	// 	}
	// }

	err := cmd.SharedActor.CheckTarget(false, false)
	if err != nil {
		return err
	}

	user, err := cmd.Config.CurrentUser()
	if err != nil {
		return err
	}

	cmd.UI.DisplayTextWithFlavor("Getting buildpacks as {{.Username}}...", map[string]interface{}{
		"Username": user.Name,
	})
	cmd.UI.DisplayNewline()

	buildpacks, warnings, err := cmd.Actor.GetBuildpacks()
	cmd.UI.DisplayWarnings(warnings)
	if err != nil {
		return err
	}

	//implement the actor layer to call the api layer to sort by position?
	//or do we need to sort here like with stacks?
	//sort.Slice(buildpacks, func(i, j int) bool { return sorting.LessIgnoreCase(stacks[i].Name, stacks[j].Name) })

	//Do we expect a no buildpack response to look like nil or an empty array of buildpacks?
	//I'm assuming empty array for now

	if len(buildpacks) == 0 {
		cmd.UI.DisplayTextWithFlavor("No buildpacks found")
	} else {
		displayTable(buildpacks, cmd.UI)
	}
	return nil
}

func displayTable(buildpacks []v7action.BuildpackTemp, display command.UI) {
	if len(buildpacks) > 0 {
		var keyValueTable = [][]string{
			{"position", "name", "stack", "enabled", "locked", "filename"},
		}
		for _, buildpack := range buildpacks {
			keyValueTable = append(keyValueTable, []string{strconv.Itoa(buildpack.Position), buildpack.Name, buildpack.Stack, strconv.FormatBool(buildpack.Enabled), strconv.FormatBool(buildpack.Locked), buildpack.Filename})
		}

		display.DisplayTableWithHeader("", keyValueTable, ui.DefaultTableSpacePadding)
	}
}
