package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kubeplay/gameserver/pkg/api/handlers"
	"github.com/kubeplay/gameserver/pkg/types"
)

type config struct {
	RegisteredAPITypes []types.Object
	Version            string
}

type Route struct {
	Path    string
	Handler func(w http.ResponseWriter, r *http.Request)
	Methods []string
}

type RouteInfo struct {
	PathPrefix  string
	Middlewares []mux.MiddlewareFunc
	SubRoutes   []Route
}

func (c *config) Routes() []RouteInfo {
	return []RouteInfo{
		{
			PathPrefix:  "/events",
			Middlewares: handlers.Event.Middlewares(),
			SubRoutes: []Route{
				{
					Path:    "",
					Handler: handlers.Event.HandlerList(),
					Methods: []string{"POST", "GET"},
				},
				{
					Path:    "/{resourceName}",
					Handler: handlers.Event.Handler(),
					Methods: []string{"GET", "DELETE", "PUT"},
				},
				{
					Path:    "/{parent}/games",
					Handler: handlers.Event.HandlerGameList(),
					Methods: []string{"POST", "GET"},
				},
				{
					Path:    "/{parent}/games/{resourceName}",
					Handler: handlers.Event.HandlerGame(),
					Methods: []string{"GET", "DELETE", "PUT"},
				},
				{
					Path:    "/{parent}/games/{resourceName}/start",
					Handler: handlers.Event.HandlerGameStart(),
					Methods: []string{"POST"},
				},
				{
					Path:    "/{parent}/games/{resourceName}/solve",
					Handler: handlers.Event.HandlerGameSolve(),
					Methods: []string{"POST"},
				},
			},
		},
		{
			PathPrefix:  "/challenges",
			Middlewares: handlers.Challenge.Middlewares(),
			SubRoutes: []Route{
				{
					Path:    "",
					Handler: handlers.Challenge.HandlerList(),
					Methods: []string{"POST", "GET"},
				},
				{
					Path:    "/{resourceName}",
					Handler: handlers.Challenge.Handler(),
					Methods: []string{"GET", "DELETE", "PUT"},
				},
			},
		},
		{
			PathPrefix:  "/players",
			Middlewares: handlers.Player.Middlewares(),
			SubRoutes: []Route{
				{
					Path:    "",
					Handler: handlers.Player.HandlerList(),
					Methods: []string{"POST", "GET"},
				},
				{
					Path:    "/{resourceName}",
					Handler: handlers.Player.Handler(),
					Methods: []string{"GET", "DELETE", "PUT"},
				},
			},
		},
		{
			PathPrefix:  "/login",
			Middlewares: handlers.Auth.Middlewares(),
			SubRoutes: []Route{
				{
					Path:    "",
					Handler: handlers.Auth.Handler(),
					Methods: []string{"GET"},
				},
			},
		},
	}
}

func (c *config) Middlewares() []mux.MiddlewareFunc {
	return []mux.MiddlewareFunc{
		authenticationMiddleware,
		decoderMiddleware,
	}
}

var Config = config{
	RegisteredAPITypes: types.RegisteredTypes,
	Version:            "/v1",
}
