package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/context"
	"github.com/gorilla/mux"
	"github.com/kubeplay/gameserver/pkg/api"
	"github.com/kubeplay/gameserver/pkg/types"
	"github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

type Handler struct{}

type Game struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Repo        string `json:"repo"`
}

type APIKey struct {
	Key   string `json:"key"`
	Token string `json:"token"`
	Email string `json:"email"`
}

func NewGame(name, desc, repo string) []byte {
	g := Game{Name: name, Description: desc, Repo: repo}
	data, _ := json.Marshal(g)
	return data
}

func GetInstance() *bolt.DB {
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		logrus.WithError(err).Fatal("failed opening file")
		// return fmt.Errorf("Failed opening)
	}
	return db
}

func GetApiKey(w http.ResponseWriter, r *http.Request, t []string) error {
	db := GetInstance()
	defer db.Close()
	var apiKey APIKey
	return db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(`apikeys`))
		if err != nil {
			return err
		}
		data := b.Get([]byte(t[0]))
		json.Unmarshal(data, &apiKey)
		if apiKey.Email == "" {
			return fmt.Errorf("failed retrieving user")
		}
		context.Set(r, "apikey", &apiKey)
		return nil
	})
}

func apiKeyAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(r.Header.Get("Authorization"), " ")
		if len(parts) != 2 || len(parts) == 2 && parts[0] != "Bearer" {
			http.Error(w, "Forbidden", http.StatusForbidden)
		} else {
			t := strings.Split(parts[1], ":")
			if len(t) != 2 {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			if err := GetApiKey(w, r, t); err != nil {
				// fmt.Sprintf("%v", err)
				http.Error(w, err.Error(), http.StatusForbidden)
				return
			}
			// db := GetInstance()

			// db.Close()

			// apiKey := &APIKey{t[0], t[1]}
			// context.Set(r, "apikey", apiKey)
			next.ServeHTTP(w, r)
		}
	})
}

// TODO: store, remove api key (POST/DELETE)
func (h *Handler) APIKeyHandler(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	ctxdata := context.Get(r, "payload")
	c, ok := ctxdata.(*types.Challenge)
	fmt.Println(c.Kind, c.Name, ok)
	// fmt.Printf("API KEY: %#v\n", ctxdata)
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		logrus.WithError(err).Fatal("failed opening file")
		// return fmt.Errorf("Failed opening)
	}
	defer db.Close()
	var data []byte
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(`challenges`))
		if b != nil {
			data = b.Get([]byte(`the-journey-begins`))
		}
		return nil
	})
	if len(data) > 0 {
		w.Header().Add("Content-Type", "application/json")
		w.Write(data)
		return
	}
	db.Update(func(tx *bolt.Tx) error {
		c, err := tx.CreateBucketIfNotExists([]byte(`challenges`))
		if err != nil {
			logrus.WithError(err).Fatal("failed retrieving bucket")
		}
		data := NewGame("foo", "A game to remember", "github.com/kubeplay/game-to-remeber")
		if err := c.Put([]byte(`the-journey-begins`), data); err != nil {
			logrus.WithError(err).Fatal("failed creating key")
		}
		w.WriteHeader(201)
		return nil
	})
	w.WriteHeader(200)
}

const dbFile = "/tmp/kubeplay.db"

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

	if err != nil {
		log.Fatalf(err.Error())
	}
	logrus.Info("Listening to 0.0.0.0:8080 ...")
	http.ListenAndServe("0.0.0.0:8080", muxr)
}
