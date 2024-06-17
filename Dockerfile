FROM golang:alpine AS builder

LABEL stage=gobuilder

RUN apk add --no-cache gcc libc-dev tzdata sqlite

WORKDIR /build

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ENV CGO_ENABLED 1

ENV GOOS linux

ENV GOARCH amd64

RUN go build -ldflags="-s -w" -o ./app/gocommies .

RUN sqlite3 ./app/commydb.sqlite < ./commies.sql


FROM alpine

RUN apk update --no-cache && apk add --no-cache ca-certificates

COPY --from=builder /usr/share/zoneinfo/Europe/Moscow  /usr/share/zoneinfo/Europe/Moscow

ENV TZ Europe/Moscow

COPY --from=builder /build/app /app

COPY ./conf.json.sample /app/conf.json

WORKDIR /app

CMD ["/app/gocommies"]
