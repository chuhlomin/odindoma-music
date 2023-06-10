// Create Spotify client and handle tracks
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/caarlos0/env"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const (
	redirectURI  = "http://localhost/spotify-callback"
	playlistName = "Один Дома «Разрыв танцполов»"
)

type config struct {
	SporifyCientID     string `env:"SPOTIFY_CLIENT_ID"`
	SporifyCientSecret string `env:"SPOTIFY_CLIENT_SECRET"`
}

var ctx = context.Background()

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

	c, err := createClient(cfg.SporifyCientID, cfg.SporifyCientSecret, "token.json")
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	user, err := c.CurrentUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}
	log.Printf("Logged in as %s\n", user.DisplayName)

	log.Println("Creating playlist...")
	p, err := c.CreatePlaylistForUser(ctx, user.ID, playlistName, "", false, false)
	if err != nil {
		return fmt.Errorf("failed to create playlist: %w", err)
	}

	log.Println("Reading tracks from file...")
	f, err := os.Open("tracks.txt")
	if err != nil {
		return fmt.Errorf("failed to open tracks file: %w", err)
	}

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if err := handleTrack(line, c, p); err != nil {
			return fmt.Errorf("failed to handle track: %w", err)
		}

	}

	return nil
}

func createClient(clientID, clientSecret, tokenFile string) (*spotify.Client, error) {
	auth := spotifyauth.New(
		spotifyauth.WithRedirectURL(redirectURI),
		spotifyauth.WithScopes(
			spotifyauth.ScopeUserReadPrivate,
			spotifyauth.ScopePlaylistReadPrivate,
			spotifyauth.ScopePlaylistModifyPrivate,
		),
		spotifyauth.WithClientID(clientID),
		spotifyauth.WithClientSecret(clientSecret),
	)

	// read token from file
	b, err := ioutil.ReadFile(tokenFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(b, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	httpClient := auth.Client(ctx, &token)

	return spotify.New(httpClient), nil
}

func handleTrack(line string, client *spotify.Client, playlist *spotify.FullPlaylist) error {
	log.Printf("Searching for track %s...", line)

	res, err := client.Search(ctx, line, spotify.SearchTypeTrack, spotify.Limit(1))
	if err != nil {
		return fmt.Errorf("failed to search for track: %w", err)
	}

	if len(res.Tracks.Tracks) == 0 {
		fmt.Printf("No tracks found for %s\n", line)
		return nil
	}

	track := res.Tracks.Tracks[0]

	log.Printf("Adding track %s...", track.Name)
	_, err = client.AddTracksToPlaylist(ctx, playlist.ID, track.ID)
	return err
}
