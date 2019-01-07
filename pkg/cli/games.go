package cli

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/kubeplay/gameserver/pkg/rest"
	"github.com/kubeplay/gameserver/pkg/store"
	"github.com/kubeplay/gameserver/pkg/types"
	"github.com/spf13/cobra"
)

func GameCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "game",
		Short:        "Create a new game",
		SilenceUsage: true,
		PreRunE:      PreLoad,
		RunE: func(cmd *cobra.Command, args []string) error {
			game := types.Game{
				TypeMeta:  types.TypeMeta{Kind: types.GameKind},
				Metadata:  types.Metadata{Name: store.NewUUID()},
				Challenge: O.Games.Challenge,
			}

			err := rest.NewRequest(nil, GameServerURL).Post().
				RequestURI("/v1/events", O.Games.Event, "games").
				Bearer(AccessToken.String()).
				Body(&game).
				Do().Into(&game)
			if err != nil {
				return err
			}
			fmt.Printf("Game %q created\n", game.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&O.Games.Challenge, "challenge", "", "Choose the challenge to create a new game.")
	cmd.Flags().StringVarP(&O.Games.Event, "event", "e", "", "The event to create the game.")
	cmd.MarkFlagRequired("challenge")
	cmd.MarkFlagRequired("event")
	return cmd
}

func GameGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "games",
		Aliases:      []string{"game"},
		PreRunE:      PreLoad,
		SilenceUsage: true,
		Short:        "Get or list specific game resources.",
		RunE: func(cmd *cobra.Command, args []string) error {
			isResourceScoped := len(args) > 0
			requestURI := path.Join("/v1/events", O.Games.Event, "games")
			if isResourceScoped {
				requestURI = path.Join(requestURI, args[0])
			}
			resp := rest.NewRequest(nil, GameServerURL).Get().
				Bearer(AccessToken.String()).
				RequestURI(requestURI).
				Do()
			if err := resp.Error(); err != nil {
				return err
			}
			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 0, 8, 2, '\t', tabwriter.AlignRight)
			defer w.Flush()
			if !isResourceScoped {
				var itemList types.GameList
				if err := resp.Into(&itemList); err != nil {
					return err
				}
				if len(itemList.Items) == 0 {
					return fmt.Errorf("No resources found.")
				}
				fmt.Fprintln(w, "NAME\tCHALLENGE\tKEYS\tDURATION\tSTATUS\t")
				for _, gm := range itemList.Items {
					startTime, _ := time.Parse(time.RFC3339, gm.Status.StartTime)
					endTime, _ := time.Parse(time.RFC3339, gm.Status.EndTime)
					delta := endTime.Sub(startTime)
					var duration string
					if gm.Status.EndTime != "" {
						duration = RoundTime(delta, time.Second).String()
					} else {
						duration = RoundTime(time.Since(startTime), time.Second).String()
					}
					if gm.Status.StartTime == "" {
						duration = "-"
					}
					completedKeys := fmt.Sprintf("%d/%d", len(gm.Status.Keys), gm.Status.RegisteredKeys)
					fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t",
						gm.Name,
						gm.Challenge,
						completedKeys,
						duration,
						gm.Status.Phase,
					)
					fmt.Fprintln(w)
				}
			} else {
				var gm types.Game
				if err := resp.Into(&gm); err != nil {
					return err
				}

				fmt.Fprintln(w, "NAME\tCHALLENGE\tKEYS\tDURATION\tSTATUS\t")
				startTime, _ := time.Parse(time.RFC3339, gm.Status.StartTime)
				endTime, _ := time.Parse(time.RFC3339, gm.Status.EndTime)
				delta := endTime.Sub(startTime)
				var duration string
				if gm.Status.EndTime != "" {
					duration = RoundTime(delta, time.Second).String()
				} else {
					duration = RoundTime(time.Since(startTime), time.Second).String()
				}
				if gm.Status.StartTime == "" {
					duration = "-"
				}
				completedKeys := fmt.Sprintf("%d/%d", len(gm.Status.Keys), gm.Status.RegisteredKeys)
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t\n",
					gm.Name,
					gm.Challenge,
					completedKeys,
					duration,
					gm.Status.Phase,
				)
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&O.Games.Event, "event", "e", "", "The event to list games.")
	cmd.MarkFlagRequired("event")
	return cmd
}

func GameSolveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "solve EVENT/GAME GAMEKEY",
		PreRunE:               PreLoad,
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return errors.New("missing required arguments: <event>/<game> <gamekey>")
			}
			if !strings.Contains(args[0], "/") {
				return errors.New("specify the resource name as <event>/<game>")
			}
			return nil
		},
		Short: "Solve the game verifying keys.",
		RunE: func(cmd *cobra.Command, args []string) error {
			parts := strings.Split(args[0], "/")
			eventName, gameName, gameKey := parts[0], parts[1], args[1]
			resp := rest.NewRequest(nil, GameServerURL).Post().
				RequestURI("/v1/events", eventName, "games", gameName, "solve").
				SetHeader(types.GameKeyHeaderName, gameKey).
				Bearer(AccessToken.String()).
				Do()
			if err := resp.Error(); err != nil {
				return err
			}
			if resp.StatusCode() == 403 {
				fmt.Println("The game key is invalid! Are you trying to hack the game? :(")
				return nil
			}
			gm := types.Game{}
			if err := resp.Into(&gm); err != nil {
				return err
			}
			gs := gm.Status.LastSolvedKey
			fmt.Printf("You've solved %q game key.\n", gs.KeyName)
			fmt.Printf("%d/%d game keys solved.\n", len(gm.Status.Keys), gm.Status.RegisteredKeys)
			if gm.Status.Phase == types.GameCompleted {
				startTime, _ := time.Parse(time.RFC3339, gm.Status.StartTime)
				elapsed := RoundTime(time.Since(startTime), time.Second).String()
				if gm.Status.StartTime == "" {
					elapsed = "-"
				}
				fmt.Printf("Congratulations! You've completed the game in %s! ;)\n", elapsed)
			}
			return nil
		},
	}
	return cmd
}

func GameStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:                   "start EVENT/GAME",
		PreRunE:               PreLoad,
		SilenceUsage:          true,
		DisableFlagsInUseLine: true,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing the resource name")
			}
			if !strings.Contains(args[0], "/") {
				return errors.New("specify the resource name as <event>/<game>")
			}
			return nil
		},
		Short: "[HOST] Start an existing game",
		RunE: func(cmd *cobra.Command, args []string) error {
			parts := strings.Split(args[0], "/")
			eventName, gameName := parts[0], parts[1]
			resp := rest.NewRequest(nil, GameServerURL).Post().
				RequestURI("/v1/events", eventName, "games", gameName, "start").
				Bearer(AccessToken.String()).
				Do()
			if err := resp.Error(); err != nil {
				return err
			}
			var gm types.Game
			if err := resp.Into(&gm); err != nil {
				return err
			}
			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 0, 8, 2, '\t', tabwriter.AlignRight)
			defer w.Flush()
			fmt.Fprintln(w, "NAME\tCHALLENGE\tKEYS\tDURATION\tSTATUS\t")
			startTime, _ := time.Parse(time.RFC3339, gm.Status.StartTime)
			endTime, _ := time.Parse(time.RFC3339, gm.Status.EndTime)
			delta := endTime.Sub(startTime)
			var duration string
			if gm.Status.EndTime != "" {
				duration = RoundTime(delta, time.Second).String()
			} else {
				duration = RoundTime(time.Since(startTime), time.Second).String()
			}
			if gm.Status.StartTime == "" {
				duration = "-"
			}
			completedKeys := fmt.Sprintf("%d/%d", len(gm.Status.Keys), gm.Status.RegisteredKeys)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t\n",
				gm.Name,
				gm.Challenge,
				completedKeys,
				duration,
				gm.Status.Phase,
			)
			return nil
		},
	}
}
