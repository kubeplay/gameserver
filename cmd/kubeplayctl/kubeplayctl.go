package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/kubeplay/gameserver/pkg/rest"
	"github.com/kubeplay/gameserver/pkg/types"
	"github.com/kubeplay/gameserver/pkg/utils"

	"github.com/kubeplay/gameserver/pkg/cli"
	"github.com/kubeplay/gameserver/pkg/version"
	"github.com/spf13/cobra"
)

// var o cli.CmdOptions

func cmd() *cobra.Command {
	create := &cobra.Command{
		Use:          "create",
		Short:        "Create game server resources.",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var obj types.Object
			switch {
			case cli.O.CreateInput == "-":
				stdin, err := ioutil.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				obj, err = utils.YamlToJson(stdin)
				if err != nil {
					return err
				}
			case strings.HasPrefix("http", cli.O.CreateInput):
				return fmt.Errorf("http input not implemented")
			default:
				fileInput, err := ioutil.ReadFile(cli.O.CreateInput)
				if err != nil {
					return err
				}
				obj, err = utils.YamlToJson(fileInput)
				if err != nil {
					return err
				}
			}
			kind := obj.GetObjectKind()
			if kind == types.GameKind {
				return fmt.Errorf("kind %q not implemented", types.GameKind)
			} else if kind == types.PolicyKind {
				kind = "policie"
			}
			restResourceKind := fmt.Sprintf("%ss", strings.ToLower(kind))
			err := rest.NewRequest(nil, cli.GameServerURL).Post().
				Bearer(cli.AccessToken.String()).
				RequestURI("v1", restResourceKind).
				Body(obj).
				Do().
				Into(obj)
			if err != nil {
				return err
			}
			meta := obj.GetObjectMeta()
			fmt.Printf("%s %q created with uid %s\n", obj.GetObjectKind(), meta.Name, meta.UID)
			return nil
		},
	}
	get := &cobra.Command{
		Use:   "get",
		Short: "Display one or many game server resources.",
	}
	del := &cobra.Command{
		Use:   "delete RESOURCE",
		Short: "Delete a resource from the game server.",
	}
	join := &cobra.Command{
		Use:   "join",
		Short: "Join into a particular event.",
	}

	root := cobra.Command{
		Use:     "kubeplay",
		Short:   "kubeplay manages the game server.",
		PostRun: cleanup,
		Run: func(cmd *cobra.Command, args []string) {
			if cli.O.ShowVersionAndExit {
				version.PrintAndExit()
			}
			cmd.Help()
		},
	}
	create.AddCommand(
		cli.GameCreateCmd(),
		cli.EventCreateCmd(),
		cli.ChallengeCreateCmd(),
		cli.PolicyCreateCmd(),
	)
	create.Flags().StringVarP(&cli.O.CreateInput, "filename", "f", "", "Filename, directory, or URL to files to use to create the resource.")
	get.AddCommand(
		cli.GameGetCmd(),
		cli.ChallengeGetCmd(),
		cli.EventGetCmd(),
		cli.PolicyGetCmd(),
	)
	del.AddCommand(
		cli.EventDeleteCmd(),
		cli.ChallengeDeleteCmd(),
	)
	join.AddCommand(cli.EventJoinCmd())
	root.AddCommand(
		create,
		del,
		get,
		join,
		cli.LoginCmd(),
		cli.GameSolveCmd(),
		cli.HackChallengeCmd(),
		cli.GameStartCmd(),
	)
	root.Flags().BoolVar(&cli.O.ShowVersionAndExit, "version", false, "Print version and exit.")
	return &root
}

func cleanup(cmd *cobra.Command, args []string) {}

func main() {
	cmd().Execute()
}
