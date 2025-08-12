package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
)

// UserData holds the mapping from a Telegram user ID to a list of visited countries.
var UserData map[int64][]string
var mutex = &sync.Mutex{}

// loadData reads the data from data.json into the UserData map.
func loadData() {
	mutex.Lock()
	defer mutex.Unlock()

	bytes, err := os.ReadFile("backend/data.json")
	if err != nil {
		if os.IsNotExist(err) {
			log.Println("data.json not found, initializing empty map.")
			UserData = make(map[int64][]string)
			return
		}
		log.Fatalf("could not read data.json: %v", err)
	}

	if err := json.Unmarshal(bytes, &UserData); err != nil {
		log.Fatalf("could not parse data.json: %v", err)
	}
	log.Println("Data loaded successfully.")
}

// saveData writes the current UserData map to data.json.
func saveData() {
	mutex.Lock()
	defer mutex.Unlock()

	bytes, err := json.MarshalIndent(UserData, "", "  ")
	if err != nil {
		log.Fatalf("could not marshal data: %v", err)
	}

	if err := os.WriteFile("backend/data.json", bytes, 0644); err != nil {
		log.Fatalf("could not write to data.json: %v", err)
	}
	log.Println("Data saved successfully.")
}

func main() {
	loadData()

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// API to get and update countries
	mux.HandleFunc("/api/countries", handleCountries)

	// Serve frontend files
	fs := http.FileServer(http.Dir("./frontend"))
	mux.Handle("/", fs)

	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", loggingMiddleware(mux)); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func handleCountries(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getCountries(w, r)
	case http.MethodPost:
		addCountry(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func getCountries(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		http.Error(w, "userId query parameter is required", http.StatusBadRequest)
		return
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid userId", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	defer mutex.Unlock()
	countries, ok := UserData[userID]
	if !ok {
		countries = []string{} // Return empty list if user not found
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(countries); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func addCountry(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID  int64  `json:"userId"`
		Country string `json:"country"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.UserID == 0 || req.Country == "" {
		http.Error(w, "userId and country are required", http.StatusBadRequest)
		return
	}

	mutex.Lock()
	UserData[req.UserID] = append(UserData[req.UserID], req.Country)
	mutex.Unlock()

	log.Printf("Saving country %s for user %d", req.Country, req.UserID)
	saveData()

	w.WriteHeader(http.StatusCreated)
}
