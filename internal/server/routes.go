package server

import (
	"encoding/json"
	"fmt"
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
	w.Header().Set("Content-Type", "application/json")

	// Insere a categoria no banco de dados e obtém o número de linhas alteradas
	rowsAffected, err := s.Db.InsertCategory(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonResp, _ := json.Marshal(map[string]string{"error": err.Error()})
		_, _ = w.Write(jsonResp)
		return
	}

	// Formata a mensagem de sucesso
	msg := fmt.Sprintf("O total de %d linha(s) foram/foi alteradas.\n", rowsAffected)

	// Converte a mensagem para JSON
	jsonResp, err := json.Marshal(map[string]string{"message": msg})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonResp, _ := json.Marshal(map[string]string{"error": "error handling JSON marshal"})
		_, _ = w.Write(jsonResp)
		return
	}

	// Escreve a resposta JSON
	_, _ = w.Write(jsonResp)
}

func (s *Server) DeleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rowsAffected, err := s.Db.DeleteCategory(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonResp, _ := json.Marshal(map[string]string{"error": "error handling JSON marshal"})
		_, _ = w.Write(jsonResp)
		return
	}

	// Formata a mensagem de sucesso
	msg := fmt.Sprintf("O total de %d linha(s) foram/foi alteradas.\n", rowsAffected)

	// Converte a mensagem para JSON
	jsonResp, err := json.Marshal(map[string]string{"message": msg})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonResp, _ := json.Marshal(map[string]string{"error": "error handling JSON marshal"})
		_, _ = w.Write(jsonResp)
		return
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
