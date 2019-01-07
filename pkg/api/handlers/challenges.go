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

var Challenge = challenge{}

type challenge struct{}

func (c *challenge) HandlerList() HandlerFn {
	return challengeListHandler
}

func (c *challenge) Handler() HandlerFn {
	return challengeHandler
}

func (c *challenge) Middlewares() []mux.MiddlewareFunc {
	return []mux.MiddlewareFunc{challengeMiddleware}
}

func challengeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ch, _ := next.(challengeHandler)
		// fmt.Printf("CH HAND: %#v\n", ch)
		logrus.WithField("method", r.Method).Info("CHALLENGE MIDDLEWARE")
		if r.Method == "POST" {
			// req := context.Get(r, "payload")
			// _, ok := req.(*types.Challenge)
			// if !ok {
			// 	http.Error(w, "unknown type found", http.StatusBadRequest)
			// 	return
			// }
			// if len(c.Keys) == 0 {
			// 	http.Error(w, "missing keys", http.StatusBadRequest)
			// 	return
			// }
			// if c.AssetsURL == "" {
			// 	http.Error(w, "missing assets url", http.StatusBadRequest)
			// 	return
			// }
		}
		next.ServeHTTP(w, r)
	})
}

func challengeHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	switch r.Method {
	// TODO: don't delete if have references to events/games
	case "DELETE":
		err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.ChallengeKind).
			Resources(strings.ToLower(types.ChallengeKind)).
			Delete(params["resourceName"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(204)
	case "GET":
		obj, err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.ChallengeKind).
			Resources(strings.ToLower(types.ChallengeKind)).
			Get(params["resourceName"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		NewResponse(w).WriteJSON(obj)
	case "PUT":
		req := context.Get(r, "payload")
		new, ok := req.(*types.Challenge)
		if !ok {
			http.Error(w, "unknown type found", http.StatusBadRequest)
			return
		}
		old, err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.ChallengeKind).
			Resources(strings.ToLower(types.ChallengeKind)).
			Get(params["resourceName"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		obj, err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.ChallengeKind).
			Resources(
				strings.ToLower(types.ChallengeKind),
				params["resourceName"],
			).Update(old, new)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		NewResponse(w).WriteJSON(obj)
	default:
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}

func challengeListHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		req := context.Get(r, "payload")
		c, ok := req.(*types.Challenge)
		if !ok {
			http.Error(w, "unknown type found", http.StatusBadRequest)
			return
		}
		resp, err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.ChallengeKind).
			Resources(strings.ToLower(types.ChallengeKind), c.Name).
			SaveObject(c)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		NewResponse(w).Status(201).WriteJSON(resp)
	case "GET":
		itemList, err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.ChallengeKind).
			Resources(strings.ToLower(types.ChallengeKind)).
			List(regexp.MustCompile(`^\/challenge`))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		items := types.ChallengeList{}
		for _, obj := range itemList {
			c := obj.(*types.Challenge)
			items.Items = append(items.Items, *c)
		}
		items.Kind = "List"
		NewResponse(w).WriteJSON(&items)
	default:
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}
