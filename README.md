# SpotifyBackupper
Automatically saves a snapshot of your spotify playlists as backup.

This Programm is meant to be run as a cronjob.

# Example using Docker

docker-compose.yaml:
```yaml
services:
  SpotifyBackupper:
    image: <image-url>
    environment:
      SPBU_SPOTIFY_CLIENT_ID: <your-client-id>
      SPBU_SPOTIFY_CLIENT_SECRET: <your-client-secret>
      SPBU_REPO_ORIGIN: https://oauth2:<github-fine-grained-token>@github.com/<username>/<backup-repo>.git
    volumes:
      - type: bind
        source: ./crontab.txt
        target: /crontab.txt
      - type: bind
        source: ./spbu_config.yaml
        target: /app/spbu_config.yaml
```

spbu_config.yaml:
```yaml
playlist_urls:
  - <some-playlist-url>
  - <some-other-playlist-url>
```

crontab.txt:
```
0 14 * * * cd /app && /app/SpotifyBackupper > /dev/stdout

```
