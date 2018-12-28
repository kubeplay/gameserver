package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/kubeplay/gameserver/pkg/rest"
	"github.com/kubeplay/gameserver/pkg/types"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

var O CmdOptions

func LoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate to the game server.",
		Run: func(cmd *cobra.Command, args []string) {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter your GitHub username/e-mail: ")
			username, _ := reader.ReadString('\n')
			fmt.Print("Enter your GitHub password or personal token: ")
			credentials, _ := terminal.ReadPassword(int(syscall.Stdin))
			fmt.Println()
			basicAuth := &rest.BasicAuth{
				Username: strings.TrimSpace(username),
				Password: strings.TrimSpace(string(credentials)),
			}
			data, err := rest.NewRequest(nil, GameServerURL).Get().
				BasicAuth(basicAuth).
				RequestURI("/v1/login").
				Do().Raw()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			var player types.PlayerClaims
			if err := json.Unmarshal(data, &player); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("Lets play %s!\n", player.Name)
			if err := WriteCredentials([]byte(player.AccessToken)); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		},
	}
	return cmd
}
