package cli

import (
	"errors"
	"fmt"
	"os"
	"path"
	"text/tabwriter"

	"github.com/kubeplay/gameserver/pkg/rest"
	"github.com/kubeplay/gameserver/pkg/types"
	"github.com/spf13/cobra"
)

// Guest
func EventGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "events",
		Aliases:      []string{"event"},
		PreRunE:      PreLoad,
		SilenceUsage: true,
		Short:        "Get or list specific event resource.",
		RunE: func(cmd *cobra.Command, args []string) error {
			isResourceScoped := len(args) > 0
			requestURI := path.Join("/v1/events")
			if isResourceScoped {
				requestURI = path.Join(requestURI, args[0])
			}
			resp := rest.NewRequest(nil, GameServerURL).Get().
				Bearer(AccessToken).
				RequestURI(requestURI).
				Do()
			w := new(tabwriter.Writer)
			w.Init(os.Stdout, 0, 8, 2, '\t', tabwriter.AlignRight)
			fmt.Fprintln(w, "NAME\tPAUSED\t")
			defer fmt.Fprintln(w)
			defer w.Flush()
			if !isResourceScoped {
				var eventList types.EventList
				if err := resp.Into(&eventList); err != nil {
					return err
				}
				for _, ev := range eventList.Items {
					fmt.Fprintf(w, "%s\t%v\t", ev.Name, ev.Paused)
				}
			} else {
				var ev types.Event
				if err := resp.Into(&ev); err != nil {
					return err
				}
				fmt.Fprintf(w, "%s\t%v\t", ev.Name, ev.Paused)
			}
			return nil
		},
	}
}

// Guest
func EventJoinCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "event",
		Short: "Join a particular event.",
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

// Host
func EventCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "event",
		SilenceUsage: true,
		Short:        "Create an event resource.",
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
