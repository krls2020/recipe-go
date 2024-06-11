package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"log/slog"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB

func init() {
	_ = godotenv.Load()

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASS"), os.Getenv("DB_NAME"))

	var errDB error
	db, errDB = sql.Open("postgres", connStr)
	if errDB != nil {
		log.Fatalf("Failed to connect to database: %v", errDB)
	}

	_, errTable := db.Exec(`CREATE TABLE IF NOT EXISTS entries (id SERIAL PRIMARY KEY, data TEXT NOT NULL);`)
	if errTable != nil {
		log.Fatalf("Failed to ensure table exists: %v", errTable)
	}
}

func main() {
	http.HandleFunc("/", addEntry)
	http.HandleFunc("/status", statusCheck)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server running on http://localhost:%s\n", port)


	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func addEntry(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method is not supported.", http.StatusNotFound)
		return
	}

	randomData := uuid.New().String()
	_, err := db.Exec(`INSERT INTO entries(data) VALUES ($1)`, randomData)
	if err != nil {
		http.Error(w, "Failed to insert entry", http.StatusInternalServerError)
		return
	}

	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM entries`).Scan(&count)
	if err != nil {
		http.Error(w, "Failed to count entries", http.StatusInternalServerError)
		return
	}

    // LOG
	log.Printf("log - entry added: %s.\n", randomData)

    slog.Info("slog.Info - entry added", "data", randomData, "total", count)
    slog.Warn("slog.Warn - entry added", "data", randomData, "total", count)
    slog.Error("slog.Error - entry added", "data", randomData, "total", count)


    response := map[string]interface{}{
        "message": `This is a simple, basic GO application running on Zerops.io, each request adds an entry to the PostgreSQL database and returns a count. See the source repository (https://github.com/zeropsio/recipe-go) for more information.`,
        "newEntry": randomData,
        "count":    count,
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(response)

}

func statusCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := map[string]string{"status": "UP"}
	json.NewEncoder(w).Encode(status)
}

