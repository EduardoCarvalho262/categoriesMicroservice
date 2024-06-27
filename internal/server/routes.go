package server

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/health", s.healthHandler)

	r.HandleFunc("/categories", s.GetAllCategoriesHandler).Methods("GET")

	r.HandleFunc("/categories", s.InsertCategoriesHandler).Methods("POST")
	r.HandleFunc("/categories/{id}", s.DeleteCategoryHandler).Methods("DELETE")

	return r
}

func (s *Server) GetAllCategoriesHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	jsonResp, err := json.Marshal(s.Db.GetAllCategories())

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) InsertCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.Db.InsertCategory(r))

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	jsonResp, err := json.Marshal(s.Db.DeleteCategory(r))

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}

func (s *Server) healthHandler(w http.ResponseWriter, _ *http.Request) {
	jsonResp, err := json.Marshal(s.Db.Health())

	if err != nil {
		log.Fatalf("error handling JSON marshal. Err: %v", err)
	}

	_, _ = w.Write(jsonResp)
}
