package server

import (
	"log"
	"net/http"

	"github.com/dt3zr/dlog/store"
	"github.com/gorilla/mux"
)

func NewHttpServer(s store.Store) {
	router := mux.NewRouter()
	router.HandleFunc("/device-ts/{macId}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		macId := vars["macId"]
		log.Printf("Request for macId = %s", macId)
		tsJson, err := s.GetTimeSeriesAsJSON(macId)
		if err == nil {
			w.Write(tsJson)
		} else {
			w.WriteHeader(404)
		}
	})
	http.ListenAndServe("localhost:8444", router)
}
