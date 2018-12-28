package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/kubeplay/gameserver/pkg/types"
	"github.com/sirupsen/logrus"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
)

var Auth = auth{}

type auth struct{}

func (c *auth) Handler() HandlerFn {
	return authHandler
}

func (c *auth) Middlewares() []mux.MiddlewareFunc {
	return []mux.MiddlewareFunc{authMiddleware}
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
			return
		}
		authData, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		context.Set(r, "github-basic-auth", string(authData))
		next.ServeHTTP(w, r)
	})
}

func authHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		basicAuth := context.Get(r, "github-basic-auth")
		req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		parts := strings.Split(basicAuth.(string), ":")
		req.SetBasicAuth(parts[0], parts[1])
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if resp.StatusCode != 200 {
			logrus.WithField("status", resp.StatusCode).Infof("failed authenticating to github")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		profile := &types.PlayerClaims{}
		if err := json.NewDecoder(resp.Body).Decode(profile); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := GenerateNewJwtToken(
			jwtSecret,
			profile,
			time.Now().UTC().Add(time.Hour*12),
		); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := json.NewEncoder(w).Encode(profile); err != nil {
			logrus.Warnf("failed encoding response %v", err)
		}
	default:
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	}
}
