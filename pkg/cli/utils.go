package cli

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/kubeplay/gameserver/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const KubeplayAddrEnv = "KUBEPLAY_ADDR"

type Token struct {
	Data []byte
}

func (t *Token) Claims() (*types.PlayerClaims, error) {
	if len(t.Data) == 0 {
		return nil, fmt.Errorf("empty access token")
	}
	parts := strings.Split(string(t.Data), ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("it's not a JWT token")
	}
	data, err := jwt.DecodeSegment(parts[1])
	if err != nil {
		return nil, err
	}
	player := types.PlayerClaims{}
	return &player, json.Unmarshal(data, &player)
}

func (t *Token) String() string {
	return string(t.Data)
}

var (
	KubePlayConfig   = os.ExpandEnv("$HOME/.kubeplay")
	KubePlayToken    = path.Join(os.ExpandEnv(KubePlayConfig), "credentials")
	AccessToken      = &Token{}
	GameServerURL, _ = url.Parse(os.Getenv("KUBEPLAY_ADDR"))
)

func PreLoad(cmd *cobra.Command, args []string) (err error) {
	if GameServerURL == nil {
		return fmt.Errorf("Wrong or missing kubeplay address %q", KubeplayAddrEnv)
	}
	AccessToken.Data, err = ioutil.ReadFile(KubePlayToken)
	AccessToken.Data = bytes.TrimSuffix(AccessToken.Data, []byte("\n"))
	return
}

// TODO: check if the token is expired
func WriteCredentials(accessToken []byte) error {
	fi, err := os.Stat(KubePlayConfig)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(KubePlayConfig, 0744); err != nil {
			return err
		}
		return ioutil.WriteFile(KubePlayToken, append(accessToken, '\n'), 0600)
	}
	switch mode := fi.Mode(); {
	case mode.IsDir():
		return ioutil.WriteFile(KubePlayToken, append(accessToken, '\n'), 0600)
	default:
		return fmt.Errorf("kubeplay config path %q is a file", KubePlayConfig)
	}
}

type CmdGames struct {
	Challenge string
	Event     string
}

type CmdOptions struct {
	ShowVersionAndExit bool

	Games CmdGames
}

func SolveGameKey(gameKeyHash, gameUID, keyName string, key types.Key) bool {
	hash := hmac.New(sha256.New, []byte(key.Value))
	hash.Write([]byte(gameUID))
	isValid := hex.EncodeToString(hash.Sum(nil)) == gameKeyHash
	logrus.WithFields(logrus.Fields{
		"name":    keyName,
		"weight":  key.Weight,
		"valid":   isValid,
		"gamekey": gameKeyHash,
	}).Info("Trying to solve game key")
	time.Sleep(1 * time.Second) // Slow compute hashes to prevent brute force hacks
	return isValid
}

func GenerateGameKey(key, gameUID string) string {
	hash := hmac.New(sha256.New, []byte(key))
	hash.Write([]byte(gameUID))
	return hex.EncodeToString(hash.Sum(nil))
}

func RoundTime(d, r time.Duration) time.Duration {
	if r <= 0 {
		return d
	}
	neg := d < 0
	if neg {
		d = -d
	}
	if m := d % r; m+m < r {
		d = d - m
	} else {
		d = d + r - m
	}
	if neg {
		return -d
	}
	return d
}
