package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/bafto/SpotifyBackupper/config"
	spotify "github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

var (
	conf   *config.Config
	ctx    context.Context
	client *spotify.Client
)

func main() {
	var err error
	ctx = context.Background()
	conf, err = config.LoadFromFile("config.json")
	if err != nil {
		log.Fatal(err)
	}

	authconfig := &oauth2.Config{
		ClientID:     conf.SPOTIFY_ID,
		ClientSecret: conf.SPOTIFY_SECRET,
	}

	client = spotify.New(authconfig.Client(ctx, &conf.Token))
	playlists, err := client.CurrentUsersPlaylists(ctx)
	if err != nil {
		log.Fatal(err)
	}
	wrappedPlaylists := make([]PlaylistWrapper, 0, playlists.Total)
	for _, playlist := range playlists.Playlists {
		wrapped, err := wrapPlaylist(playlist)
		if err != nil {
			log.Println(err)
			continue
		}
		wrappedPlaylists = append(wrappedPlaylists, wrapped)
	}
	file, err := json.MarshalIndent(wrappedPlaylists, "", "\t")
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(conf.BackupPath, file, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
}

func GetAllTracksInPlaylist(playlistID spotify.ID) ([]spotify.PlaylistTrack, error) {
	tracks, err := client.GetPlaylistTracks(ctx, playlistID)
	if err != nil {
		return nil, err
	}
	result := make([]spotify.PlaylistTrack, 0, tracks.Total)
	result = append(result, tracks.Tracks...)
	for offset := len(tracks.Tracks); offset < tracks.Total; offset = offset + len(tracks.Tracks) {
		tracks, err = client.GetPlaylistTracks(ctx, playlistID, spotify.Offset(offset))
		if err != nil {
			return nil, err
		}
		result = append(result, tracks.Tracks...)
	}
	return result, nil
}

type TrackWrapper struct {
	Name    string   `json:"name"`
	Artists []string `json:"artists"`
	AddedAt string   `json:"addedat"`
}

type PlaylistWrapper struct {
	Name   string         `json:"name"`
	Tracks []TrackWrapper `json:"tracks"`
}

func wrapPlaylist(playlist spotify.SimplePlaylist) (result PlaylistWrapper, err error) {
	result.Name = playlist.Name
	tracks, err := GetAllTracksInPlaylist(playlist.ID)
	if err != nil {
		return result, err
	}
	result.Tracks = make([]TrackWrapper, 0, len(tracks))
	for i, track := range tracks {
		result.Tracks = append(result.Tracks, TrackWrapper{
			Name:    track.Track.Name,
			Artists: make([]string, 0, len(track.Track.Artists)),
			AddedAt: track.AddedAt,
		})
		for _, artist := range track.Track.Artists {
			result.Tracks[i].Artists = append(result.Tracks[i].Artists, artist.Name)
		}
	}
	return result, nil
}
