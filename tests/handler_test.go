package tests

import (
	"categoriesMicroservice/internal/database"
	"categoriesMicroservice/internal/server"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	s := &server.Server{}

	NewServer := &server.Server{
		Port: port,

		Db: database.New(),
	}
	newServer := httptest.NewServer(http.HandlerFunc(s.GetAllCategoriesHandler))
	defer newServer.Close()
	resp, err := http.Get(newServer.URL)
	if err != nil {
		t.Fatalf("error making request to server. Err: %v", err)
	}
	defer resp.Body.Close()
	// Assertions
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}
	expected := "[\n    {\n        \"id\": 1,\n        \"nome\": \"Item A\",\n        \"quantidade\": 15.75,\n        \"limite\": 200\n    },\n]"
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body. Err: %v", err)
	}
	if expected != string(body) {
		t.Errorf("expected response body to be %v; got %v", expected, string(body))
	}
}
