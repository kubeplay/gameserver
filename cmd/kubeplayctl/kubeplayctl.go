package main

import (
	"github.com/kubeplay/gameserver/pkg/cli"
	"github.com/kubeplay/gameserver/pkg/version"
	"github.com/spf13/cobra"
)

// var o cli.CmdOptions

func cmd() *cobra.Command {
	create := &cobra.Command{
		Use:   "create",
		Short: "Create game server resources.",
	}
	get := &cobra.Command{
		Use:   "get",
		Short: "Display one or many game server resources.",
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
	)
	get.AddCommand(
		cli.GameGetCmd(),
		cli.ChallengeGetCmd(),
		cli.EventGetCmd(),
	)
	join.AddCommand(cli.EventJoinCmd())

	root.AddCommand(
		create,
		get,
		join,
		cli.LoginCmd(),
		cli.GameSolveCmd(),
		cli.HackChallengeCmd(),
	)
	root.Flags().BoolVar(&cli.O.ShowVersionAndExit, "version", false, "Print version and exit.")
	return &root
}

func cleanup(cmd *cobra.Command, args []string) {}

func main() {
	cmd().Execute()
}
