FROM golang:alpine as build

COPY . /app
WORKDIR /app
RUN go build -ldflags "-s -w" -o SpotifyBackupper .

FROM alpine as run

COPY --from=build /app/SpotifyBackupper .
COPY crontab.txt /crontab.txt
RUN crontab /crontab.txt

CMD [ "crond", "-f", "-l", "5" ]
