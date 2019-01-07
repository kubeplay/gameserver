package handlers

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"

	"github.com/kubeplay/gameserver/pkg/store"
	"github.com/kubeplay/gameserver/pkg/types"
)

var Policy = policy{}

type policy struct{}

func (c *policy) HandlerList() HandlerFn {
	return policyListHandler
}

func (c *policy) Handler() HandlerFn {
	return policyHandler
}

func (c *policy) Middlewares() []mux.MiddlewareFunc {
	return []mux.MiddlewareFunc{}
}

func policyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func policyHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	switch r.Method {
	case "DELETE":
		err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.PolicyKind).
			Resources(strings.ToLower(types.PolicyKind)).
			Delete(params["resourceName"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(204)
	case "GET":
		obj, err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.PolicyKind).
			Resources(strings.ToLower(types.PolicyKind)).
			Get(params["resourceName"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		NewResponse(w).WriteJSON(obj)
	default:
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}

func policyListHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		req := context.Get(r, "payload")
		p, ok := req.(*types.Policy)
		if !ok {
			http.Error(w, "unknown type found", http.StatusBadRequest)
			return
		}
		resp, err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.PolicyKind).
			Resources(strings.ToLower(types.PolicyKind), p.Name).
			SaveObject(p)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		NewResponse(w).Status(201).WriteJSON(resp)
	case "GET":
		itemList, err := store.New(dbConfig.file, dbConfig.bucket).
			Kind(types.PolicyKind).
			Resources(strings.ToLower(types.PolicyKind)).
			List(regexp.MustCompile(`^\/policy`))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		items := types.PolicyList{}
		for _, obj := range itemList {
			c := obj.(*types.Policy)
			items.Items = append(items.Items, *c)
		}
		items.Kind = "List"
		NewResponse(w).WriteJSON(&items)
	default:
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}
