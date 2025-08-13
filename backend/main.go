package main

import (
	"bytes"
	"encoding/json"
	"image/color"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/fogleman/gg"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	geojson "github.com/paulmach/go.geojson"
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

func generateMapImage(visitedCountries []string) (*bytes.Buffer, error) {
	// Load and parse the GeoJSON file
	raw, err := ioutil.ReadFile("data/countries.geo.json")
	if err != nil {
		return nil, err
	}

	fc, err := geojson.UnmarshalFeatureCollection(raw)
	if err != nil {
		return nil, err
	}

	// Create a new image context
	const (
		width  = 1024
		height = 512
	)
	dc := gg.NewContext(width, height)
	dc.SetRGB(0.9, 0.9, 0.9) // Set background color
	dc.Clear()

	// Create a map for quick lookup of visited countries
	visitedSet := make(map[string]bool)
	for _, country := range visitedCountries {
		visitedSet[country] = true
	}

	// Find the bounding box of the world to scale the map
	minX, minY, maxX, maxY := 180.0, 90.0, -180.0, -90.0
	for _, feature := range fc.Features {
		if feature.Geometry == nil {
			continue
		}
		for _, polygon := range feature.Geometry.Polygon {
			for _, point := range polygon {
				if point[0] < minX {
					minX = point[0]
				}
				if point[0] > maxX {
					maxX = point[0]
				}
				if point[1] < minY {
					minY = point[1]
				}
				if point[1] > maxY {
					maxY = point[1]
				}
			}
		}
	}

	scaleX := float64(width) / (maxX - minX)
	scaleY := float64(height) / (maxY - minY)

	// Draw each country
	for _, feature := range fc.Features {
		if feature.Properties == nil {
			continue
		}
		countryName, ok := feature.Properties["name"].(string)
		if !ok {
			continue
		}
		isVisited := visitedSet[countryName]

		if isVisited {
			dc.SetColor(color.RGBA{R: 212, G: 172, B: 13, A: 255}) // Gold for visited
		} else {
			dc.SetColor(color.RGBA{R: 200, G: 200, B: 200, A: 255}) // Gray for not visited
		}

		// Handle both Polygon and MultiPolygon geometries
		if feature.Geometry.IsPolygon() {
			for _, ring := range feature.Geometry.Polygon {
				drawPolygon(dc, ring, minX, maxY, scaleX, scaleY)
			}
		} else if feature.Geometry.IsMultiPolygon() {
			for _, polygon := range feature.Geometry.MultiPolygon {
				for _, ring := range polygon {
					drawPolygon(dc, ring, minX, maxY, scaleX, scaleY)
				}
			}
		}

		dc.Fill()
	}

	// Encode the image to a buffer
	buffer := new(bytes.Buffer)
	if err := dc.EncodePNG(buffer); err != nil {
		return nil, err
	}

	return buffer, nil
}

func drawPolygon(dc *gg.Context, polygon [][]float64, minX, maxY, scaleX, scaleY float64) {
	if len(polygon) == 0 {
		return
	}
	for i, point := range polygon {
		x := (point[0] - minX) * scaleX
		y := (maxY - point[1]) * scaleY
		if i == 0 {
			dc.MoveTo(x, y)
		} else {
			dc.LineTo(x, y)
		}
	}
	dc.ClosePath()
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

	go startTelegramBot()

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
					photo.Caption = "Here is your map of visited countries!"
					if _, err := bot.Send(photo); err != nil {
						log.Printf("Error sending photo: %v", err)
					}
				}
			}
		default:
			msg.Text = "I don't know that command."
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Error sending message: %v", err)
			}
		}
	}
}
