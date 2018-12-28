package cli

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"text/tabwriter"

	"github.com/kubeplay/gameserver/pkg/rest"
	"github.com/kubeplay/gameserver/pkg/types"
	"github.com/spf13/cobra"
)

// Guest
func ChallengeGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "challenges",
		Aliases:      []string{"challenge"},
		PreRunE:      PreLoad,
		SilenceUsage: true,
		Short:        "Get or list specific challenge resource.",
		RunE: func(cmd *cobra.Command, args []string) error {
			isResourceScoped := len(args) > 0
			requestURI := path.Join("/v1/challenges")
			if isResourceScoped {
				requestURI = path.Join(requestURI, args[0])
			}
			resp := rest.NewRequest(nil, GameServerURL).Get().
				Bearer(AccessToken).
				RequestURI(requestURI).
				Do()
			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 0, 8, 2, '\t', tabwriter.AlignRight)
			fmt.Fprintln(w, "NAME\tKEYS\t")
			defer fmt.Fprintln(w)
			defer w.Flush()
			if !isResourceScoped {
				var itemList types.ChallengeList
				if err := resp.Into(&itemList); err != nil {
					return err
				}
				for _, c := range itemList.Items {
					fmt.Fprintf(w, "%s\t%d\t", c.Name, len(c.Keys))
				}
			} else {
				var c types.Challenge
				if err := resp.Into(&c); err != nil {
					return err
				}
				fmt.Fprintf(w, "%s\t%d\t", c.Name, len(c.Keys))
			}
			return nil
		},
	}
}

// Host
func ChallengeCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "challenge",
		Short: "TODO: Create a new challenge.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing the resource name")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
}

func HackChallengeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "hack EVENT/GAME",
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
		Short: "Hack all challenge keys generating game keys.",
		RunE: func(cmd *cobra.Command, args []string) error {
			parts := strings.Split(args[0], "/")
			eventName, gameName := parts[0], parts[1]
			var gm types.Game
			err := rest.NewRequest(nil, GameServerURL).Get().
				RequestURI("/v1/events", eventName, "games", gameName).
				Bearer(AccessToken.String()).
				Do().Into(&gm)
			if err != nil {
				return err
			}
			var c types.Challenge
			err = rest.NewRequest(nil, GameServerURL).Get().
				RequestURI("/v1/challenges", gm.Challenge).
				Bearer(AccessToken.String()).
				Do().Into(&c)
			if err != nil {
				return err
			}
			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 0, 8, 2, '\t', tabwriter.AlignRight)
			fmt.Fprintln(w, "KEYNAME\tWEIGHT\tCHALLENGE\tHASH\t")
			defer w.Flush()
			for keyName, key := range c.Keys {
				gameKeyHash := GenerateGameKey(key.Value, gm.UID)
				fmt.Fprintf(w, "%s\t%.1f\t%s\t%s\t",
					keyName,
					key.Weight,
					c.Name,
					gameKeyHash,
				)
				fmt.Fprintln(w)
			}
			return nil
		},
	}
	return cmd
}
