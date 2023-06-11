// Grab token from Spotify API and save to file
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"

	"github.com/caarlos0/env"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

const redirectURI = "http://localhost/spotify-callback"

type config struct {
	SpotifyCientID     string `env:"SPOTIFY_CLIENT_ID"`
	SpotifyCientSecret string `env:"SPOTIFY_CLIENT_SECRET"`
}

func main() {
	log.Println("Starting...")

	if err := run(); err != nil {
		log.Fatalf("Error: %v", err)
	}

	log.Println("Done.")
}

func run() error {
	var cfg config
	if err := env.Parse(&cfg); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}

	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopePlaylistReadPrivate,
			spotifyauth.ScopePlaylistModifyPrivate,
		),
		spotifyauth.WithClientID(cfg.SpotifyCientID),
		spotifyauth.WithClientSecret(cfg.SpotifyCientSecret),
	)

	// set `state` var to something random
	state := randomString(16)

	url := auth.AuthURL(state)

	fmt.Println("Please log in to Spotify by visiting the following page in your browser:", url)

	// start HTTP server, listen for callback, stop HTTP server
	ctx := context.Background()

	http.HandleFunc("/spotify-callback", func(w http.ResponseWriter, r *http.Request) {
		tok, err := auth.Token(ctx, state, r)
		if err != nil {
			log.Printf("Couldn't get token: %v", err)
			http.Error(w, "Couldn't get token", http.StatusForbidden)
			return
		}

		// encode token as JSON and save to file
		f, err := os.OpenFile("token.json", os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("Couldn't open file: %v", err)
			http.Error(w, "Couldn't open file", http.StatusForbidden)
			return
		}

		err = json.NewEncoder(f).Encode(tok)
		if err != nil {
			log.Printf("Couldn't encode token: %v", err)
			http.Error(w, "Couldn't encode token", http.StatusForbidden)
			return
		}

		log.Println("Token saved to file")
		os.Exit(0)
	})

	err := http.ListenAndServe(":80", nil)
	if err != nil {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
