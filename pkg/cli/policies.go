package cli

import (
	"errors"
	"fmt"
	"os"
	"path"
	"text/tabwriter"

	"gopkg.in/yaml.v2"

	"github.com/kubeplay/gameserver/pkg/rest"
	"github.com/kubeplay/gameserver/pkg/types"
	"github.com/kubeplay/gameserver/pkg/utils"
	"github.com/spf13/cobra"
)

func PolicyGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "policies",
		Aliases:      []string{"policy"},
		PreRunE:      PreLoad,
		SilenceUsage: true,
		Short:        "Get or list policies.",
		RunE: func(cmd *cobra.Command, args []string) error {
			isResourceScoped := len(args) > 0
			requestURI := path.Join("/v1/policies")
			if isResourceScoped {
				requestURI = path.Join(requestURI, args[0])
			}
			resp := rest.NewRequest(nil, GameServerURL).Get().
				Bearer(AccessToken).
				RequestURI(requestURI).
				Do()
			if err := resp.Error(); err != nil {
				return err
			}
			if !isResourceScoped {
				w := new(tabwriter.Writer)
				w.Init(os.Stdout, 0, 8, 2, '\t', tabwriter.AlignRight)
				defer w.Flush()
				fmt.Fprintln(w, "NAME\tAGE\t")
				var policyList types.PolicyList
				if err := resp.Into(&policyList); err != nil {
					return err
				}
				for _, p := range policyList.Items {
					d := utils.GetDeltaDuration(p.CreatedAt, "")
					fmt.Fprintf(w, "%s\t%s\t\n", p.Name, d)
				}
			} else {
				var p types.Policy
				if err := resp.Into(&p); err != nil {
					return err
				}
				data, err := yaml.Marshal(p)
				if err != nil {
					return err
				}
				fmt.Print(string(data))
			}
			return nil
		},
	}
}

func PolicyCreateCmd() *cobra.Command {
	return &cobra.Command{
		Use:          "policy",
		SilenceUsage: true,
		Short:        "Create a policy resource.",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("missing the resource name")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			p := &types.Policy{
				TypeMeta: types.TypeMeta{Kind: types.PolicyKind},
				Metadata: types.Metadata{Name: args[0]},
			}
			err := rest.NewRequest(nil, GameServerURL).Post().
				Bearer(AccessToken.String()).
				RequestURI("/v1/policies").
				Body(p).
				Do().
				Into(p)
			if err != nil {
				return err
			}
			fmt.Printf("Policy %q created with uid %s\n", p.Name, p.UID)
			return nil
		},
	}
}

// func EventDeleteCmd() *cobra.Command {
// 	return &cobra.Command{
// 		Use:          "events",
// 		Aliases:      []string{"event"},
// 		PreRunE:      PreLoad,
// 		SilenceUsage: true,
// 		Short:        "Get or list specific event resource.",
// 		Args: func(cmd *cobra.Command, args []string) error {
// 			if len(args) < 1 {
// 				return errors.New("missing the resource name")
// 			}
// 			return nil
// 		},
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			requestURI := path.Join("/v1/events", args[0])
// 			_, err := rest.NewRequest(nil, GameServerURL).Delete().
// 				Bearer(AccessToken).
// 				RequestURI(requestURI).
// 				Do().Raw()
// 			if err != nil {
// 				return err
// 			}
// 			fmt.Printf("Event %q deleted!\n", args[0])
// 			return nil
// 		},
// 	}
// }
