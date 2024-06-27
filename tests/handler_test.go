package tests

import (
	"bytes"
	"categoriesMicroservice/internal/database"
	"categoriesMicroservice/internal/model"
	"categoriesMicroservice/internal/server"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

var db *sql.DB

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "11",
		Env: []string{
			"POSTGRES_PASSWORD=secret",
			"POSTGRES_USER=user_name",
			"POSTGRES_DB=postgres",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://user_name:secret@%s/postgres?sslmode=disable", hostAndPort)

	log.Println("Connecting to database on url: ", databaseUrl)

	resource.Expire(120) // Tell docker to hard kill the container in 120 seconds

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 120 * time.Second
	if err = pool.Retry(func() error {
		db, err = sql.Open("pgx", databaseUrl)
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// Create the database
	_, err = db.Exec(`CREATE DATABASE category`)
	if err != nil {
		log.Fatalf("Could not create database: %s", err)
	}

	// Reconnect to the newly created database
	databaseUrl = fmt.Sprintf("postgres://user_name:secret@%s/category?sslmode=disable", hostAndPort)
	db, err = sql.Open("pgx", databaseUrl)
	if err != nil {
		log.Fatalf("Could not connect to new database: %s", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("Could not ping new database: %s", err)
	}

	// Create table and insert test data
	_, err = db.Exec(`
		CREATE TABLE category (
			id SERIAL PRIMARY KEY,
			nome VARCHAR(100),
			quantidade DECIMAL(10, 2),
			limite INT
		)
	`)
	if err != nil {
		log.Fatalf("Could not create table: %s", err)
	}

	_, err = db.Exec(`
		INSERT INTO category (nome, quantidade, limite) 
		VALUES ('Item A', 15.75, 200)
	`)
	if err != nil {
		log.Fatalf("Could not insert test data: %s", err)
	}

	// run tests
	code := m.Run()

	// Purge the container after tests
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	os.Exit(code)
}

func TestGetAllCategoriesHandler(t *testing.T) {
	// Use the db initialized in TestMain
	s := &server.Server{Db: database.NewTest(db)}
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
	expected := `[{"id":1,"nome":"Item A","quantidade":15.75,"limite":200}]`
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body. Err: %v", err)
	}
	if expected != string(body) {
		t.Errorf("expected response body to be %v; got %v", expected, string(body))
	}
}

func TestInsertCategoriesHandler(t *testing.T) {
	// Use the db initialized in TestMain
	s := &server.Server{Db: database.NewTest(db)}
	newServer := httptest.NewServer(http.HandlerFunc(s.InsertCategoriesHandler))
	defer newServer.Close()

	request := model.Category{Nome: "Teste", Quantidade: 15.35, Limite: 300}

	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(request)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.Post(newServer.URL, "application/json", &buf)
	if err != nil {
		t.Fatalf("error making request to server. Err: %v", err)
	}
	defer resp.Body.Close()
	// Assertions
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}
	var expected = "{\"message\":\"O total de 1 linha(s) foram/foi alteradas.\\n\"}"
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body. Err: %v", err)
	}
	if expected != string(body) {
		t.Errorf("expected response body to be %v; got %v", expected, string(body))
	}
}

func TestDeleteCategoriesHandler(t *testing.T) {
	// Use the db initialized in TestMain
	s := &server.Server{Db: database.NewTest(db)}
	// Configure o roteador para que a rota DELETE seja tratada
	router := mux.NewRouter()
	router.HandleFunc("/category/{id}", s.DeleteCategoryHandler).Methods(http.MethodDelete)
	newServer := httptest.NewServer(router)
	defer newServer.Close()

	// Faça uma solicitação DELETE para deletar uma categoria específica
	req, err := http.NewRequest(http.MethodDelete, newServer.URL+"/category/1", nil)
	if err != nil {
		t.Fatalf("error creating request. Err: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("error making request to server. Err: %v", err)
	}
	defer resp.Body.Close()

	// Assertions
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}

	expected := "{\"message\":\"O total de 1 linha(s) foram/foi alteradas.\\n\"}"
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body. Err: %v", err)
	}

	if expected != string(body) {
		t.Errorf("expected response body to be %v; got %v", expected, string(body))
	}
}
