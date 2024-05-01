package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

// Stage represents a single stage item
type Stage struct {
	ID        int               `json:"id"`
	StageName string            `json:"stage_name"`
	Stages    map[string]string `json:"stages"`
}

var db *sql.DB

func main() {
	// Initialize database connection
	var err error
	db, err = sql.Open("postgres", "postgres://tavito:mamacita@localhost:5432/stages?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	// Initialize Gorilla Mux router
	r := mux.NewRouter()

	// Define API endpoints
	r.HandleFunc("/stages", getStages).Methods("GET")
	r.HandleFunc("/stages/{id}", getStageByID).Methods("GET")
	r.HandleFunc("/stages", createStage).Methods("POST")
	r.HandleFunc("/stages/{id}", updateStage).Methods("PUT")
	r.HandleFunc("/stages/{id}", deleteStage).Methods("DELETE")

	// Serve static files from the "static" directory
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Add CORS middleware
	r.Use(corsMiddleware)

	log.Println("Server started on port 8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}

// Middleware function to add CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// getStages retrieves all stages from the database
func getStages(w http.ResponseWriter, r *http.Request) {
	// Query to fetch all stages from the database
	rows, err := db.Query("SELECT id, stage_name, stages FROM stages")
	if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Initialize a slice to store retrieved stages
	var allStages []Stage

	// Iterate through query results and append each stage to the slice
	for rows.Next() {
		var stage Stage
		var stagesJSON []byte
		if err := rows.Scan(&stage.ID, &stage.StageName, &stagesJSON); err != nil {
			log.Println("Error scanning row:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Unmarshal the JSON-encoded stages string into a map[string]string
		if err := json.Unmarshal(stagesJSON, &stage.Stages); err != nil {
			log.Println("Error decoding stages JSON:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		allStages = append(allStages, stage)
	}
	if err := rows.Err(); err != nil {
		log.Println("Error iterating rows:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Encode the slice of stages as JSON and write it to the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(allStages); err != nil {
		log.Println("Error encoding JSON:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// getStageByID retrieves a specific stage by ID from the database
func getStageByID(w http.ResponseWriter, r *http.Request) {
	// Extract stage ID from URL parameters
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	// Query to fetch stage by ID from the database
	row := db.QueryRow("SELECT stage_name, stages FROM stages WHERE id = $1", id)

	// Initialize a Stage struct to store the retrieved stage
	var stage Stage
	var stagesJSON []byte

	// Scan the query result into the Stage struct
	err = row.Scan(&stage.StageName, &stagesJSON)
	if err == sql.ErrNoRows {
		http.Error(w, "Stage not found", http.StatusNotFound)
		return
	} else if err != nil {
		log.Println("Error querying database:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Unmarshal the JSON-encoded stages string into a map[string]string
	if err := json.Unmarshal(stagesJSON, &stage.Stages); err != nil {
		log.Println("Error decoding stages JSON:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Encode the retrieved stage as JSON and write it to the response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stage); err != nil {
		log.Println("Error encoding JSON:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// createStage creates a new stage in the database
func createStage(w http.ResponseWriter, r *http.Request) {
	// Parse request body to get the new stage details
	var stage Stage
	err := json.NewDecoder(r.Body).Decode(&stage)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	// Convert the stages map to a JSON-encoded string
	stagesJSON, err := json.Marshal(stage.Stages)
	if err != nil {
		log.Println("Error encoding stages map to JSON:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Insert the new stage into the database
	_, err = db.Exec("INSERT INTO stages (stage_name, stages) VALUES ($1, $2)", stage.StageName, stagesJSON)
	if err != nil {
		log.Println("Error inserting into database:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set the response status code to 201 Created
	w.WriteHeader(http.StatusCreated)
}

// updateStage updates an existing stage in the database
func updateStage(w http.ResponseWriter, r *http.Request) {
	// Extract stage ID from URL parameters
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	// Parse request body to get the updated stage details
	var updatedStage Stage
	err = json.NewDecoder(r.Body).Decode(&updatedStage)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	// Convert the stages map to a JSON-encoded string
	stagesJSON, err := json.Marshal(updatedStage.Stages)
	if err != nil {
		log.Println("Error encoding stages map to JSON:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Update the existing stage in the database
	_, err = db.Exec("UPDATE stages SET stage_name = $1, stages = $2 WHERE id = $3", updatedStage.StageName, stagesJSON, id)
	if err != nil {
		log.Println("Error updating database:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set the response status code to 200 OK
	w.WriteHeader(http.StatusOK)
}

// deleteStage deletes a stage from the database
func deleteStage(w http.ResponseWriter, r *http.Request) {
	// Extract stage ID from URL parameters
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "Invalid stage ID", http.StatusBadRequest)
		return
	}

	// Delete the stage from the database
	_, err = db.Exec("DELETE FROM stages WHERE id = $1", id)
	if err != nil {
		log.Println("Error deleting from database:", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

/*
///////////Postgres Database//////////
psql

\l

CREATE DATABASE stages;

DROP DATABASE stages;     //for deleting a database

\c stages

pwd

\i /Users/tavito/Documents/go/vocabulary-builder-stages-API/stages.sql

\dt


////////////////Curl Commands///////////////////

curl -X GET http://localhost:8000/stages

curl -X GET http://localhost:8000/stages/{id}

curl -X POST \
  http://localhost:8000/stages \
  -H 'Content-Type: application/json' \
  -d '{
	"stage_name": "IELTS",
	"stages": {
		"listening": "http://localhost:8080/dataset/category/listening",
		"reading": "http://localhost:8080/dataset/category/reading",
		"writing": "http://localhost:8080/dataset/category/writing",
		"speaking": "http://localhost:8080/dataset/category/speaking"
	}
}'

curl -X POST \
  http://localhost:8000/stages \
  -H 'Content-Type: application/json' \
  -d '{
	"stage_name": "IELTS",
	"stages": {
	  "easy-stage": "http://localhost:8080/dataset/category/easy-word",
  	  "hard-stage": "http://localhost:8080/dataset/category/hard-word"
	}
}'

curl -X PUT \
  http://localhost:8000/stages/2 \
  -H 'Content-Type: application/json' \
  -d '{
	"stage_name": "TOEFL",
	"stages": {
	  "easy-stage": "http://localhost:8080/dataset/category/easy-word",
  	  "hard-stage": "http://localhost:8080/dataset/category/hard-word"
	}
}'

curl -X DELETE http://localhost:8000/stages/{id}

*/
