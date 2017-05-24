#!/bin/bash
if test -z "$TG_APP_ID"; then
    echo "** Please set env variables from .env before building."
    exit 1
fi
if test -z "$TG_API_HASH"; then
    echo "** Please set env variables from .env before building."
    exit 1
fi

VER="$(cat VERSION)"

LDFLAGS="-X main.version=$VER -X main.apiID=$TG_APP_ID -X main.apiHash=$TG_API_HASH"

go build -ldflags "$LDFLAGS" ./cmd/telegram-exporter 
GOOS=windows GOARCH=386 go build -ldflags "$LDFLAGS" ./cmd/telegram-exporter 
