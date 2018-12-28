package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/kubeplay/gameserver/pkg/cli"

	"github.com/sirupsen/logrus"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/kubeplay/gameserver/pkg/store"
	"github.com/kubeplay/gameserver/pkg/types"
)

func (c *event) HandlerGameList() HandlerFn {
	return gameListHandler
}

func (c *event) HandlerGame() HandlerFn {
	return gameHandler
}

func (c *event) HandlerGameStart() HandlerFn {
	return gameStartHandler
}

func (c *event) HandlerGameSolve() HandlerFn {
	return gameSolveHandler
}

func gameHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	switch r.Method {
	case "GET":
		obj, err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.GameKind).
			Resources(
				strings.ToLower(types.EventKind),
				params["parent"],
				strings.ToLower(types.GameKind),
			).Get(params["resourceName"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		NewResponse(w).WriteJSON(obj)
	default:
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}
func gameStartHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	switch r.Method {
	case "POST":
		s := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.GameKind).
			Resources(
				strings.ToLower(types.EventKind),
				params["parent"],
				strings.ToLower(types.GameKind),
			)
		obj, err := s.Get(params["resourceName"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		gm := obj.(*types.Game)
		if gm.Status.Phase == types.GameRunning {
			NewResponse(w).WriteJSON(obj)
			return
		}
		gm.Status.StartTime = time.Now().UTC().Format(time.RFC3339)
		gm.Status.Phase = types.GameRunning
		obj, err = s.Update(gm, gm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		NewResponse(w).WriteJSON(obj)
	default:
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}

func gameSolveHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	switch r.Method {
	case "POST":
		gameKeyHash := r.Header.Get(types.GameKeyHeaderName)
		if gameKeyHash == "" {
			msg := fmt.Sprintf("%q header not set or empty", types.GameKeyHeaderName)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
		obj, err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.EventKind).
			Resources(strings.ToLower(types.EventKind)).
			Get(params["parent"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// The event must be active to approve keys
		ev := obj.(*types.Event)
		if ev.Paused {
			http.Error(w, "The event isn't active", http.StatusBadRequest)
			return
		}
		obj, err = store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.GameKind).
			Resources(
				strings.ToLower(types.EventKind),
				params["parent"],
				strings.ToLower(types.GameKind),
			).Get(params["resourceName"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		// The game must be running to approve keys
		gm := obj.(*types.Game)
		if gm.Status.Phase != types.GameRunning {
			http.Error(w, "The game isn't running", http.StatusBadRequest)
			return
		}
		obj, err = store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.ChallengeKind).
			Resources(strings.ToLower(types.ChallengeKind)).
			Get(gm.Challenge)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		chl := obj.(*types.Challenge)
		for keyName, key := range chl.Keys {
			if ok := cli.SolveGameKey(gameKeyHash, gm.UID, keyName, key); ok {
				for _, status := range gm.Status.Keys {
					if status.KeyName == keyName {
						logrus.WithField("key", keyName).Warn("Key already validated, noop")
						NewResponse(w).WriteJSON(gm)
						return
					}
				}
				gameStatus := types.GameKeyStatus{
					KeyName:    keyName,
					Approved:   true,
					ApprovedAt: time.Now().UTC().Format(time.RFC3339),
					Weight:     key.Weight,
				}
				gm.Status.Keys = append(gm.Status.Keys, gameStatus)
				gm.Status.LastSolvedKey = gameStatus
				// All keys are validated, means the player completed the game!
				if len(chl.Keys) == len(gm.Status.Keys) {
					gm.Status.Phase = types.GameCompleted
					gm.Status.EndTime = time.Now().UTC().Format(time.RFC3339)
				}
				obj, err = store.New(dbConfig.file, dbConfig.bucket).
					Kind(types.GameKind).
					Resources(
						strings.ToLower(types.EventKind),
						params["parent"],
						strings.ToLower(types.GameKind),
						gm.Name,
					).Update(gm, gm)
				NewResponse(w).WriteJSON(gm)
				return
			}
		}
		http.Error(w, "Key not validated", http.StatusForbidden)
	default:
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}

func gameListHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	switch r.Method {
	case "POST":
		req := context.Get(r, "payload")
		gm, ok := req.(*types.Game)
		if !ok {
			http.Error(w, "unknown type found", http.StatusBadRequest)
			return
		}
		obj := context.Get(r, "player")
		pl, ok := obj.(*types.PlayerClaims)
		if !ok {
			http.Error(w, "missing player from context", http.StatusBadRequest)
			return
		}
		obj, err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.ChallengeKind).
			Resources(strings.ToLower(types.ChallengeKind)).
			Get(gm.Challenge)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		c := obj.(*types.Challenge)
		_, err = store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.EventKind).
			Resources(strings.ToLower(types.EventKind)).
			Get(params["parent"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		logrus.WithFields(logrus.Fields{
			"event":     params["parent"],
			"player":    pl.Username(),
			"challenge": c.Name,
		}).Infof("Creating a new game %q", gm.Name)
		gm.Status = types.GameStatus{
			Phase:          types.GamePending,
			RegisteredKeys: len(c.Keys),
		}
		gm.Player = pl.Username()
		resp, err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.GameKind).
			Resources(
				strings.ToLower(types.EventKind),
				params["parent"],
				strings.ToLower(types.GameKind),
				gm.Name,
			).SaveObject(gm)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		NewResponse(w).Status(201).WriteJSON(resp)
	case "GET":
		items, err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.GameKind).
			Resources(
				strings.ToLower(types.EventKind),
				params["parent"],
				strings.ToLower(types.GameKind),
			).List(regexp.MustCompile(`^\/event\/[a-z0-9-]+\/game`))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		itemList := types.GameList{}
		for _, obj := range items {
			gm := obj.(*types.Game)
			itemList.Items = append(itemList.Items, *gm)
		}
		itemList.Kind = "List"
		NewResponse(w).WriteJSON(&itemList)

	default:
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}

// func validateGameKey(gameKeyHash, gameUID, keyName string, key types.Key) bool {
// 	hash := hmac.New(sha256.New, []byte(key.Value))
// 	hash.Write([]byte(gameUID))
// 	isValid := hex.EncodeToString(hash.Sum(nil)) == gameKeyHash
// 	logrus.WithFields(logrus.Fields{
// 		"name":    keyName,
// 		"weight":  key.Weight,
// 		"valid":   isValid,
// 		"gamekey": gameKeyHash,
// 	}).Info("Trying to solve game key")
// 	time.Sleep(1 * time.Second) // Slow compute hashes to prevent brute force hacks
// 	return isValid
// }
