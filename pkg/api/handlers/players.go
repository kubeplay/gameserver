package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/kubeplay/gameserver/pkg/store"
	"github.com/kubeplay/gameserver/pkg/types"
	"github.com/sirupsen/logrus"
)

var Player = player{}

type player struct{}

func (c *player) HandlerList() HandlerFn {
	return playerListHandler
}

func (c *player) Handler() HandlerFn {
	return playerHandler
}

func (c *player) Middlewares() []mux.MiddlewareFunc {
	return []mux.MiddlewareFunc{playerMiddleware}
}

func playerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.WithField("method", r.Method).Info("PLAYER MIDDLEWARE")
		next.ServeHTTP(w, r)
	})
}

func playerHandler(w http.ResponseWriter, r *http.Request) {}
func playerListHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		req := context.Get(r, "payload")
		pl, ok := req.(*types.Player)
		if !ok {
			http.Error(w, "unknown type found", http.StatusBadRequest)
			return
		}
		resp, err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.PlayerKind).
			Resources(strings.ToLower(types.PlayerKind), pl.Name).
			SaveObject(pl)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		NewResponse(w).Status(201).WriteJSON(resp)
	case "GET":
		items, err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.PlayerKind).
			Resources(strings.ToLower(types.PlayerKind)).
			List(regexp.MustCompile(`^\/player`))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		itemList := types.PlayerList{}
		for _, obj := range items {
			pl := obj.(*types.Player)
			itemList.Items = append(itemList.Items, *pl)
		}
		itemList.Kind = "List"
		NewResponse(w).WriteJSON(&itemList)
	default:
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}
