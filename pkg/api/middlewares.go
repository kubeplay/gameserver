package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/context"
	"github.com/kubeplay/gameserver/pkg/api/handlers"
	"github.com/kubeplay/gameserver/pkg/types"
	"github.com/sirupsen/logrus"
)

func authenticationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.WithField("method", r.Method).Info("AUTHENTICATION MIDDLEWARE")
		if strings.HasSuffix(r.RequestURI, "/v1/login") {
			next.ServeHTTP(w, r)
			return
		}
		parts := strings.Split(r.Header.Get("Authorization"), " ")
		if len(parts) != 2 || len(parts) == 2 && parts[0] != "Bearer" {
			// fmt.Printf("PARTS %#v\n", r.Header.Get("Authorization"))
			logrus.WithField("method", r.Method).Infof("Missing Authorization header - 403 / not implemented yet!")
			// FORBIDDEN

			// http.Error(w, "Forbidden", http.StatusForbidden)
			next.ServeHTTP(w, r)
			return
		}
		switch t := parts[1]; {
		case strings.HasPrefix(t, "gamekey:"):
			logrus.WithField("key", t).Infof("Game Key Token")
			// remove the header "gamekey:"
			context.Set(r, "gamekey", t[8:])
		case strings.HasPrefix(t, "token-"):
			// API Key Token
			logrus.Infof("API Based token, not implemented yet")
		default:
			pl, err := handlers.DecodeUserToken(parts[1], os.Getenv("JWT_SECRET"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			context.Set(r, "player", pl)
		}
		next.ServeHTTP(w, r)

	})
}

func decoderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.WithField("method", r.Method).Info("GLOBAL MIDDLEWARE")
		switch r.Method {
		case "POST", "PUT", "PATCH":
			payload, err := ioutil.ReadAll(r.Body)
			if err != nil {
				msg := fmt.Sprintf("failed reading body: %v", err)
				http.Error(w, msg, http.StatusBadRequest)
				return
			}
			if len(payload) == 0 {
				next.ServeHTTP(w, r)
				return
			}
			typeMeta := &types.TypeMeta{}
			if err := json.Unmarshal(payload, typeMeta); err != nil {
				msg := fmt.Sprintf("failed decoding to type meta: %v", err)
				http.Error(w, msg, http.StatusBadRequest)
				return
			}
			obj, err := types.Decode(typeMeta, payload)
			if err != nil {
				msg := fmt.Sprintf("failed decoding object %v: %v", typeMeta.Kind, err)
				http.Error(w, msg, http.StatusBadRequest)
				return
			}
			context.Set(r, "payload", obj)
			next.ServeHTTP(w, r)
		default:
			next.ServeHTTP(w, r)
		}
	})
}
