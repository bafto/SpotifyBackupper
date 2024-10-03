package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/spf13/viper"
	spotify "github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2/clientcredentials"
)

func configure() {
	viper.SetDefault("spotify_client_id", "")
	viper.SetDefault("spotify_client_secret", "")
	viper.SetDefault("log_level", "INFO")
	viper.SetDefault("timeout", time.Second*10)
	viper.SetDefault("playlist_urls", []string{})
	viper.SetDefault("file_prefix", "spbu_backup_")

	viper.SetEnvPrefix("SPBU")
	viper.AutomaticEnv()

	viper.AddConfigPath(".")
	viper.SetConfigName("spbu_config")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

// returns the access token
func authenticate(ctx context.Context, clientContext context.Context) (*spotify.Client, error) {
	config := &clientcredentials.Config{
		ClientID:     viper.GetString("spotify_client_id"),
		ClientSecret: viper.GetString("spotify_client_secret"),
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	httpClient := spotifyauth.New().Client(clientContext, token)
	return spotify.New(httpClient), nil
}

func main() {
	var err error
	ctx := context.Background()

	configure()

	slog.Info("authenticating")
	authCtx, cancelAuthCtx := context.WithTimeout(ctx, viper.GetDuration("timeout"))
	defer cancelAuthCtx()
	client, err := authenticate(authCtx, ctx)
	if err != nil {
		slog.Error("authentication failed", "err", err)
		return
	}

	var playlists []*spotify.FullPlaylist
	for _, url := range viper.GetStringSlice("playlist_urls") {
		slog.Info("getting playlist", "url", url)
		playlistID, err := getPlaylistIdFromURL(url)
		if err != nil {
			slog.Warn("unable to parse playlist ID from url", "url", url, "err", err)
			continue
		}
		playlist, err := client.GetPlaylist(ctx, spotify.ID(playlistID))
		if err != nil {
			slog.Warn("unable to get playlist data", "playlist-id", url, "err", err)
			continue
		}

		playlists = append(playlists, playlist)
	}

	wrappedPlaylists := make([]PlaylistWrapper, 0, len(playlists))
	for _, playlist := range playlists {
		slog.Info("wrapping playlist", "playlist-id", playlist.ID)
		wrapped, err := wrapPlaylist(ctx, client, playlist)
		if err != nil {
			slog.Warn("unable to wrap playlist", "playlist-id", playlist.ID, "err", err)
			continue
		}
		wrappedPlaylists = append(wrappedPlaylists, wrapped)
	}

	file, err := json.MarshalIndent(wrappedPlaylists, "", "\t")
	if err != nil {
		slog.Error("marshal json error", "err", err)
		return
	}
	file_path := viper.GetString("file_prefix") + time.Now().Format("2006-01-02_15-04-05") + ".json"
	slog.Info("writing to file", "path", file_path)
	err = os.WriteFile(file_path, file, os.ModePerm)
	if err != nil {
		slog.Error("write file error", "err", err)
	}
}

type ItemWrapper struct {
	ID      string   `json:"id"`
	Name    string   `json:"name"`
	Artists []string `json:"artists"`
	AddedAt string   `json:"addedat"`
}

type PlaylistWrapper struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Items []ItemWrapper `json:"tracks"`
}

func wrapPlaylist(ctx context.Context, client *spotify.Client, playlist *spotify.FullPlaylist) (result PlaylistWrapper, err error) {
	result.Name = playlist.Name
	playlistItems, err := getAllPlaylistItems(ctx, client, playlist.ID)
	if err != nil {
		slog.Warn("unable to get playlist Items", "playlist-id", playlist.ID, "err", err)
		return PlaylistWrapper{}, err
	}
	return PlaylistWrapper{
		ID:   string(playlist.ID),
		Name: playlist.Name,
		Items: map_slice(playlistItems, func(item spotify.PlaylistItem) ItemWrapper {
			return ItemWrapper{
				ID:   string(item.Track.Track.ID),
				Name: item.Track.Track.Name,
				Artists: map_slice(item.Track.Track.Artists, func(artist spotify.SimpleArtist) string {
					return artist.Name
				}),
				AddedAt: item.AddedAt,
			}
		}),
	}, nil
}

func getAllPlaylistItems(ctx context.Context, client *spotify.Client, playlistId spotify.ID) ([]spotify.PlaylistItem, error) {
	page, err := client.GetPlaylistItems(ctx, playlistId)
	if err != nil {
		return nil, err
	}
	items := make([]spotify.PlaylistItem, 0, page.Total)
	items = append(items, page.Items...)
	for {
		err = client.NextPage(ctx, page)
		items = append(items, page.Items...)
		if err == spotify.ErrNoMorePages {
			return items, nil
		}
		if err != nil {
			return items, err
		}
	}
}

func getPlaylistIdFromURL(playlistUrl string) (string, error) {
	parsed, err := url.Parse(playlistUrl)
	if err != nil {
		return "", err
	}
	return path.Base(parsed.Path), nil
}

func map_slice[T, R any](s []T, f func(T) R) []R {
	result := make([]R, 0, len(s))
	for i := range s {
		result = append(result, f(s[i]))
	}
	return result
}
