package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"gopkg.in/mgo.v2"
)

type contextKey struct {
	name string
}

// Server es el servidor de la API.
type Server struct {
	db *mgo.Session
}

var contextKeyAPIKey = &contextKey{"api-key"}

func APIKey(ctx context.Context) (string, bool) {
	key, ok := ctx.Value(contextKeyAPIKey).(string)
	return key, ok
}

func main() {
	var (
		addr  = flag.String("addr", ":8080", "endpoint address")
		mongo = flag.String("mongo", "localhost", "mongodb address")
	)
	log.Println("Conectando a Mongo", *mongo)
	db, err := mgo.Dial(*mongo)
	if err != nil {
		log.Fatalln("No se pudo conectar a Mongo:", err)
	}
	defer db.Close()
	s := &Server{
		db: db,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/polls/",
		withCORS(withAPIKey(s.handlePolls)))
	log.Println("Iniciando servidor web en ", *addr)
	http.ListenAndServe(":8080", mux)
	log.Println("Deteniendo...")
}

func withAPIKey(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.Query().Get("key")
		if !isValidAPIKey(key) {
			respondErr(w, r, http.StatusUnauthorized, "Llave API inválida")
			return
		}
		ctx := context.WithValue(r.Context(),
			contextKeyAPIKey, key)
		fn(w, r.WithContext(ctx))
	}
}

func isValidAPIKey(key string) bool {
	return key == "abc123"
}

func withCORS(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Expose-Headers",
			"Location")
		fn(w, r)
	}
}
