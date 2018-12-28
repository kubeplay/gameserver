package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/kubeplay/gameserver/pkg/store"

	"github.com/sirupsen/logrus"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/kubeplay/gameserver/pkg/types"
)

var Event = event{}

type event struct{}

func (c *event) HandlerList() HandlerFn {
	return eventListHandler
}

func (c *event) Handler() HandlerFn {
	return eventHandler
}

func (c *event) Middlewares() []mux.MiddlewareFunc {
	return []mux.MiddlewareFunc{eventMiddleware}
}

func eventMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.WithField("method", r.Method).Infof("EVENT/GAMES MIDDLEWARE")
		next.ServeHTTP(w, r)
	})
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	switch r.Method {
	case "GET":
		obj, err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.EventKind).
			Resources(strings.ToLower(types.EventKind)).
			Get(params["resourceName"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		NewResponse(w).WriteJSON(obj)
	case "DELETE":
		err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.EventKind).
			Resources(strings.ToLower(types.EventKind)).
			Delete(params["resourceName"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(204)
	default:
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}
func eventListHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		req := context.Get(r, "payload")
		ev, ok := req.(*types.Event)
		if !ok {
			http.Error(w, "unknown type found", http.StatusBadRequest)
			return
		}
		resp, err := store.New(dbConfig.file, dbConfig.bucket).
			Resources(strings.ToLower(types.EventKind), ev.Name).
			SaveObject(ev)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		NewResponse(w).Status(201).WriteJSON(resp)
	case "GET":
		items, err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.EventKind).
			Resources(strings.ToLower(types.EventKind)).
			List(regexp.MustCompile(`\/event\/[a-z0-9-]+$`))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		itemList := types.EventList{}
		for _, obj := range items {
			ev := obj.(*types.Event)
			itemList.Items = append(itemList.Items, *ev)
		}
		itemList.Kind = "List"
		NewResponse(w).WriteJSON(&itemList)
	default:
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}
