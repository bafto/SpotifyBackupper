FROM golang:alpine as build

COPY . /app
WORKDIR /app
RUN go build -ldflags "-s -w" -o SpotifyBackupper .

FROM alpine as run

RUN apk add --no-cache git
RUN apk add --no-cache tzdata

COPY --from=build /app/SpotifyBackupper /app/SpotifyBackupper

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

CMD [ "/entrypoint.sh" ]
