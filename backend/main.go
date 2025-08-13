package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

func deleteCountry(w http.ResponseWriter, r *http.Request) {
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

	countries, ok := UserData[req.UserID]
	if !ok {
		mutex.Unlock()
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	newCountries := []string{}
	found := false
	for _, country := range countries {
		if country != req.Country {
			newCountries = append(newCountries, country)
		} else {
			found = true
		}
	}

	if !found {
		mutex.Unlock()
		http.Error(w, "Country not found in user's list", http.StatusNotFound)
		return
	}

	UserData[req.UserID] = newCountries
	mutex.Unlock()

	log.Printf("Deleting country %s for user %d", req.Country, req.UserID)
	saveData()

	w.WriteHeader(http.StatusOK)
}
func startTelegramBot() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Println("TELEGRAM_BOT_TOKEN not set, skipping bot initialization.")
		return
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Printf("Error initializing bot: %v", err)
		return
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message updates
			continue
		}

		if !update.Message.IsCommand() {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		switch update.Message.Command() {
		case "map":
			userID := update.Message.From.ID
			mutex.Lock()
			countries, ok := UserData[userID]
			mutex.Unlock()

			if !ok || len(countries) == 0 {
				msg.Text = "You haven't added any countries yet. Use the web app to add some!"
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Error sending message: %v", err)
				}
			} else {
				// Generate the map image
				photoBytes, err := generateMapImage(countries)
				if err != nil {
					log.Printf("Error generating map image: %v", err)
					msg.Text = "Sorry, I couldn't generate your map."
					if _, err := bot.Send(msg); err != nil {
						log.Printf("Error sending message: %v", err)
					}
				} else {
					photo := tgbotapi.NewPhoto(update.Message.Chat.ID, tgbotapi.FileBytes{
						Name:  "map.png",
						Bytes: photoBytes.Bytes(),
					})
					photo.Caption = fmt.Sprintf("@%s here is your map of %d visited countries!", update.Message.From.UserName, len(countries))
					if _, err := bot.Send(photo); err != nil {
						log.Printf("Error sending photo: %v", err)
					}
				}
			}
		case "list":
			userID := update.Message.From.ID
			mutex.Lock()
			countries, ok := UserData[userID]
			mutex.Unlock()

			if !ok || len(countries) == 0 {
				msg.Text = "You haven't added any countries yet. Use the web app to add some!"
			} else {
				var countryList strings.Builder
				for _, country := range countries {
					countryList.WriteString("- " + country + ",\n")
				}

				msg.Text = fmt.Sprintf("@%s here is your list:\n%s", update.Message.From.UserName, countryList.String())
			}
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Error sending message: %v", err)
			}
		default:
			msg.Text = "I don't know that command."
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Error sending message: %v", err)
			}
		}
	}
}
