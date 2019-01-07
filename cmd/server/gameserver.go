package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/kubeplay/gameserver/pkg/api"
	"github.com/kubeplay/gameserver/pkg/api/auth"
	"github.com/sirupsen/logrus"
)

func main() {
	muxr := mux.NewRouter()
	root := muxr.PathPrefix("/v1").Subrouter()
	for _, r := range api.Config.Routes() {
		if r.PathPrefix == "" {
			for _, sr := range r.SubRoutes {
				s := root.PathPrefix(sr.Path).Subrouter()
				s.HandleFunc("", sr.Handler).Methods(sr.Methods...)
				s.Use(r.Middlewares...)
			}
			continue
		}
		s := root.PathPrefix(r.PathPrefix).Subrouter()
		s.Use(r.Middlewares...)
		for _, sr := range r.SubRoutes {
			s.HandleFunc(sr.Path, sr.Handler).Methods(sr.Methods...)
		}
	}
	// Global Middleware
	// root.Use(api.Config.Middlewares()...)
	muxr.Use(api.Config.Middlewares()...)
	err := muxr.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			methods, _ := route.GetMethods()
			// if len(methods) == 0 {
			// 	return nil
			// }
			logrus.WithFields(logrus.Fields{
				"methods": strings.Join(methods, ","),
			}).Info(pathTemplate)
		}
		return nil
	})

	err = auth.NewUserPolicy(api.UserPoliciesFile, "github|sandromello", false)
	if err != nil {
		log.Fatalf(err.Error())
	}

	if err != nil {
		log.Fatalf(err.Error())
	}
	logrus.Info("Listening to 0.0.0.0:8080 ...")
	http.ListenAndServe("0.0.0.0:8080", muxr)
}
