package main

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type contextKey string

const userContextKey = contextKey("user")

type TelegramUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// validateTelegramAuth checks the integrity of the data received from the Telegram Web App.
func validateTelegramAuth(initData string, token string) (*TelegramUser, error) {
	// The data is received as a query string, so we need to parse it.
	parsed, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse initData: %w", err)
	}

	hash := parsed.Get("hash")
	if hash == "" {
		return nil, fmt.Errorf("hash parameter not found in initData")
	}

	// The user data is in the 'user' field, JSON-encoded.
	userJSON := parsed.Get("user")
	if userJSON == "" {
		return nil, fmt.Errorf("user parameter not found in initData")
	}

	// To validate, we need to re-create the data-check-string.
	// This is a string of all key-value pairs from initData, sorted alphabetically,
	// in the format "key=value", joined by a newline character.
	var dataCheck []string
	for k, v := range parsed {
		if k != "hash" {
			dataCheck = append(dataCheck, fmt.Sprintf("%s=%s", k, v[0]))
		}
	}
	sort.Strings(dataCheck)
	dataCheckString := strings.Join(dataCheck, "\n")

	// The secret key for HMAC is the bot token, prefixed with "WebAppData".
	secretKey := hmac.New(sha256.New, []byte("WebAppData"))
	secretKey.Write([]byte(token))

	// Calculate the HMAC signature.
	mac := hmac.New(sha256.New, secretKey.Sum(nil))
	mac.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac.Sum(nil))

	// Compare the hashes.
	if expectedHash != hash {
		return nil, fmt.Errorf("invalid hash")
	}

	// Additionally, we can check the auth_date to prevent replay attacks.
	authDateStr := parsed.Get("auth_date")
	if authDateStr == "" {
		return nil, fmt.Errorf("auth_date not found")
	}
	authTimestamp, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid auth_date format")
	}
	authDate := time.Unix(authTimestamp, 0)

	if time.Since(authDate) > 24*time.Hour { // Or a much shorter duration
		return nil, fmt.Errorf("authentication data is outdated")
	}

	// If everything is fine, decode the user data.
	var user TelegramUser
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data: %w", err)
	}

	return &user, nil
}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		initData := r.Header.Get("X-Telegram-Init-Data")
		if initData == "" {
			http.Error(w, "Unauthorized: X-Telegram-Init-Data header missing", http.StatusUnauthorized)
			return
		}

		token := os.Getenv("TELEGRAM_BOT_TOKEN")
		if token == "" {
			http.Error(w, "Internal Server Error: TELEGRAM_BOT_TOKEN not configured", http.StatusInternalServerError)
			return
		}

		user, err := validateTelegramAuth(initData, token)
		if err != nil {
			log.Printf("Auth validation failed: %v", err)
			http.Error(w, fmt.Sprintf("Forbidden: %v", err), http.StatusForbidden)
			return
		}

		// Add user to the request context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
