package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/filipkroca/revgeo"
	"github.com/fogleman/gg"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	geojson "github.com/paulmach/go.geojson"

	"github.com/korjavin/countrycounter/backend/store"
)

// geoJSONPath is the path to the countries GeoJSON file, relative to the working directory.
// Override in tests since test working directory is the package directory (backend/).
var geoJSONPath = "data/countries.geo.json"

// server bundles the HTTP handlers around their dependency: the visits repo.
type server struct {
	repo *store.VisitsRepo
}

// mercatorY converts a latitude in degrees to the Mercator Y coordinate.
// Latitude is clamped to [-85, 85] to avoid infinity at the poles.
func mercatorY(lat float64) float64 {
	if lat > 85.0 {
		lat = 85.0
	} else if lat < -85.0 {
		lat = -85.0
	}
	latRad := lat * math.Pi / 180.0
	return math.Log(math.Tan(math.Pi/4 + latRad/2))
}

func generateMapImage(visitedCountries []string) (*bytes.Buffer, error) {
	// Load and parse the GeoJSON file
	raw, err := os.ReadFile(geoJSONPath)
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
	dc.SetRGB(0.65, 0.81, 0.89) // Ocean blue background
	dc.Clear()

	// Create a map for quick lookup of visited countries
	visitedSet := make(map[string]bool)
	for _, country := range visitedCountries {
		visitedSet[country] = true
	}

	// Find the bounding box of the world to scale the map using Mercator Y
	minX, maxX := 180.0, -180.0
	minMercY, maxMercY := math.MaxFloat64, -math.MaxFloat64

	updateBounds := func(points [][]float64) {
		for _, point := range points {
			if len(point) < 2 {
				continue
			}
			lon, lat := point[0], point[1]
			if lon < minX {
				minX = lon
			}
			if lon > maxX {
				maxX = lon
			}
			my := mercatorY(lat)
			if my < minMercY {
				minMercY = my
			}
			if my > maxMercY {
				maxMercY = my
			}
		}
	}

	for _, feature := range fc.Features {
		if feature.Geometry == nil {
			continue
		}
		if feature.Geometry.IsPolygon() {
			for _, ring := range feature.Geometry.Polygon {
				updateBounds(ring)
			}
		} else if feature.Geometry.IsMultiPolygon() {
			for _, polygon := range feature.Geometry.MultiPolygon {
				for _, ring := range polygon {
					updateBounds(ring)
				}
			}
		}
	}

	if maxX <= minX || maxMercY <= minMercY {
		return nil, fmt.Errorf("could not compute valid map bounds from GeoJSON")
	}

	scaleX := float64(width) / (maxX - minX)
	scaleY := float64(height) / (maxMercY - minMercY)

	// Draw each country
	for _, feature := range fc.Features {
		if feature.Properties == nil {
			continue
		}
		countryName, ok := feature.Properties["name"].(string)
		if !ok {
			continue
		}
		if feature.Geometry == nil {
			continue
		}
		isVisited := visitedSet[countryName]

		if isVisited {
			dc.SetColor(color.RGBA{R: 231, G: 76, B: 60, A: 255}) // Warm red for visited
		} else {
			dc.SetColor(color.RGBA{R: 230, G: 230, B: 220, A: 255}) // Warm white for unvisited
		}

		// Handle both Polygon and MultiPolygon geometries
		if feature.Geometry.IsPolygon() {
			for _, ring := range feature.Geometry.Polygon {
				drawPolygon(dc, ring, minX, maxMercY, scaleX, scaleY)
			}
		} else if feature.Geometry.IsMultiPolygon() {
			for _, polygon := range feature.Geometry.MultiPolygon {
				for _, ring := range polygon {
					drawPolygon(dc, ring, minX, maxMercY, scaleX, scaleY)
				}
			}
		}

		dc.FillPreserve()
		dc.SetColor(color.RGBA{R: 100, G: 100, B: 100, A: 180})
		dc.SetLineWidth(0.5)
		dc.Stroke()
	}

	// Encode the image to a buffer
	buffer := new(bytes.Buffer)
	if err := dc.EncodePNG(buffer); err != nil {
		return nil, err
	}

	return buffer, nil
}

func drawPolygon(dc *gg.Context, polygon [][]float64, minX, maxMercY, scaleX, scaleY float64) {
	if len(polygon) == 0 {
		return
	}
	for i, point := range polygon {
		if len(point) < 2 {
			continue
		}
		x := (point[0] - minX) * scaleX
		y := (maxMercY - mercatorY(point[1])) * scaleY
		if i == 0 {
			dc.MoveTo(x, y)
		} else {
			dc.LineTo(x, y)
		}
	}
	dc.ClosePath()
}

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "backend/data.db"
	}

	db, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("could not open database at %s: %v", dbPath, err)
	}
	defer db.Close()

	if err := store.Migrate(db); err != nil {
		log.Fatalf("could not run migrations: %v", err)
	}

	repo := store.NewVisitsRepo(db)

	const jsonPath = "backend/data.json"
	imported, err := MaybeImportJSON(repo, jsonPath)
	if err != nil {
		log.Fatalf("auto-import from %s failed: %v", jsonPath, err)
	}
	switch {
	case imported > 0:
		log.Printf("Auto-imported %d rows from %s", imported, jsonPath)
	default:
		// MaybeImportJSON returned 0 for one of two reasons; check which so
		// operators can distinguish "DB already had data" from "no JSON file
		// to import."
		if _, statErr := os.Stat(jsonPath); statErr == nil {
			log.Printf("DB already populated, ignoring %s", jsonPath)
		} else {
			log.Printf("No %s found, starting with current DB state", jsonPath)
		}
	}

	srv := &server{repo: repo}

	go startTelegramBot(repo)

	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// API to get and update countries
	mux.HandleFunc("/api/countries", srv.handleCountries)

	// Serve only the GeoJSON file (not the entire data directory for security)
	mux.HandleFunc("/data/countries.geo.json", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./data/countries.geo.json")
	})

	// Serve frontend files
	fs := http.FileServer(http.Dir("./frontend"))
	mux.Handle("/", fs)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, loggingMiddleware(mux)); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func (s *server) handleCountries(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getCountries(w, r)
	case http.MethodPost:
		s.addCountry(w, r)
	case http.MethodDelete:
		s.deleteCountry(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *server) getCountries(w http.ResponseWriter, r *http.Request) {
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

	countries, err := s.repo.List(userID)
	if err != nil {
		log.Printf("repo.List failed for user %d: %v", userID, err)
		http.Error(w, "Failed to load countries", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(countries); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (s *server) addCountry(w http.ResponseWriter, r *http.Request) {
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

	if len(req.Country) > 100 {
		http.Error(w, "country name too long", http.StatusBadRequest)
		return
	}

	if err := s.repo.Add(req.UserID, req.Country); err != nil {
		log.Printf("repo.Add failed for user %d / %s: %v", req.UserID, req.Country, err)
		http.Error(w, "Failed to save country", http.StatusInternalServerError)
		return
	}

	log.Printf("Saving country %s for user %d", req.Country, req.UserID)
	w.WriteHeader(http.StatusCreated)
}

func (s *server) deleteCountry(w http.ResponseWriter, r *http.Request) {
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

	if len(req.Country) > 100 {
		http.Error(w, "country name too long", http.StatusBadRequest)
		return
	}

	removed, err := s.repo.Delete(req.UserID, req.Country)
	if err != nil {
		log.Printf("repo.Delete failed for user %d / %s: %v", req.UserID, req.Country, err)
		http.Error(w, "Failed to delete country", http.StatusInternalServerError)
		return
	}
	if !removed {
		http.Error(w, "Country not found for this user", http.StatusNotFound)
		return
	}

	log.Printf("Deleting country %s for user %d", req.Country, req.UserID)
	w.WriteHeader(http.StatusOK)
}

// getCountryFromLocation converts latitude/longitude to a country name
// Returns the country name if found, or an error if the location cannot be geocoded
func getCountryFromLocation(latitude, longitude float64) (string, error) {
	decoder := revgeo.Decoder{}

	// revgeo.Geocode expects (lng, lat) in GeoJSON order
	isoCode, err := decoder.Geocode(longitude, latitude)
	if err != nil {
		return "", fmt.Errorf("failed to geocode location: %w", err)
	}

	// Convert ISO code to full country name
	countryName, ok := ISOToCountryName[isoCode]
	if !ok {
		return "", fmt.Errorf("unknown country code: %s", isoCode)
	}

	return countryName, nil
}

// geocodeLocation is the package-level seam that handleLocation uses to look up
// the country for a (lat, lng). Tests override it to avoid hitting the real
// reverse-geocoder.
var geocodeLocation = getCountryFromLocation

// handleLocation runs the bot's location flow: geocode → check duplicate → add
// → recompute count for the reply. The returned reply text is always
// non-empty and is what the bot should send back to the user. The error is
// non-nil only for unexpected infrastructure failures and is intended for
// logging — the user-facing reply already explains the situation.
func handleLocation(repo *store.VisitsRepo, userID int64, lat, lng float64) (string, error) {
	country, err := geocodeLocation(lat, lng)
	if err != nil {
		return "Sorry, I couldn't determine the country from that location. Please make sure you're sharing a location within a country's borders.", err
	}

	alreadyVisited, err := repo.Has(userID, country)
	if err != nil {
		return "Sorry, something went wrong saving that country. Please try again.", fmt.Errorf("repo.Has: %w", err)
	}
	if alreadyVisited {
		return fmt.Sprintf("You've already added %s to your list! 🗺️", country), nil
	}

	if err := repo.Add(userID, country); err != nil {
		return "Sorry, something went wrong saving that country. Please try again.", fmt.Errorf("repo.Add: %w", err)
	}

	countries, err := repo.List(userID)
	if err != nil {
		// The country was added; only the count lookup failed. Send a
		// success reply without the count rather than a misleading error.
		return fmt.Sprintf("Added %s to your visited countries! 🎉", country), fmt.Errorf("repo.List: %w", err)
	}

	return fmt.Sprintf("Added %s to your visited countries! 🎉\nYou've now visited %d countries. Use /map to see your progress!", country, len(countries)), nil
}

func startTelegramBot(repo *store.VisitsRepo) {
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
		if update.Message.From == nil { // channel posts and similar — skip
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		// Handle location messages
		if update.Message.Location != nil {
			userID := update.Message.From.ID
			lat := update.Message.Location.Latitude
			lng := update.Message.Location.Longitude

			log.Printf("Received location from user %d: lat=%f, lng=%f", userID, lat, lng)

			reply, err := handleLocation(repo, userID, lat, lng)
			if err != nil {
				log.Printf("handleLocation failed for user %d: %v", userID, err)
			}
			msg.Text = reply
			if _, err := bot.Send(msg); err != nil {
				log.Printf("Error sending message: %v", err)
			}
			continue
		}

		// Handle command messages
		if !update.Message.IsCommand() {
			continue
		}

		switch update.Message.Command() {
		case "map":
			userID := update.Message.From.ID
			countries, err := repo.List(userID)
			if err != nil {
				log.Printf("repo.List failed for user %d: %v", userID, err)
				msg.Text = "Sorry, I couldn't load your countries."
				if _, err := bot.Send(msg); err != nil {
					log.Printf("Error sending message: %v", err)
				}
				continue
			}

			if len(countries) == 0 {
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
			countries, err := repo.List(userID)
			if err != nil {
				log.Printf("repo.List failed for user %d: %v", userID, err)
				msg.Text = "Sorry, I couldn't load your countries."
			} else if len(countries) == 0 {
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
		case "suggest":
			userID := update.Message.From.ID
			visitedCountries, err := repo.List(userID)
			if err != nil {
				log.Printf("repo.List failed for user %d: %v", userID, err)
				visitedCountries = nil
			}

			suggestions := GetCountrySuggestions(visitedCountries, 8)
			msg.Text = FormatSuggestions(suggestions)

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
