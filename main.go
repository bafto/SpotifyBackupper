package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/bafto/SpotifyBackupper/git"
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
	viper.SetDefault("repo_origin", "")
	viper.SetDefault("git_user_name", "")
	viper.SetDefault("git_user_email", "")

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

	slog.Info("setting git user info")
	if err := git.ConfigureUser(ctx, viper.GetString("git_user_name"), viper.GetString("git_user_email")); err != nil {
		slog.Warn("failed to configure git user", "err", err)
	}

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
		slog.Info("wrapped playlist", "playlist-id", wrapped.ID)
	}

	file, err := json.MarshalIndent(wrappedPlaylists, "", "\t")
	if err != nil {
		slog.Error("marshal json error", "err", err)
		return
	}

	repo_origin := viper.GetString("repo_origin")
	repo_url, err := url.Parse(repo_origin)
	if err != nil {
		slog.Error("failed to parse repo_origin", "err", err, "origin", repo_origin)
	}
	repo_name := strings.TrimSuffix(filepath.Base(repo_url.Path), filepath.Ext(repo_url.Path))

	slog.Info("checking/cloning repo")
	if err := git.CreateRepoIfNotExists(ctx, repo_name, repo_origin); err != nil {
		slog.Error("failed to initialized git repo", "err", err)
		return
	}

	file_path := repo_name + "/spbu_backup.json"
	slog.Info("writing to file", "path", file_path)
	if err = os.WriteFile(file_path, file, os.ModePerm); err != nil {
		slog.Error("write file error", "err", err)
		return
	}

	slog.Info("committing and pushing changes")
	if err := git.CommitAndPushChanges(ctx, repo_name, time.Now().Format("backup 2006-01-02_15-04-05")); err != nil {
		slog.Error("commit and push error", "err", err)
		return
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
		if err == spotify.ErrNoMorePages {
			return items, nil
		}
		if err != nil {
			return items, err
		}
		items = append(items, page.Items...)
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
