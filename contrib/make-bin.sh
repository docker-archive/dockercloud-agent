#!/bin/bash

VERSION=$(cat VERSION)
DEST=$1
GOOS=${GOOS:-linux} GOARCH=${GOARCH:-amd64} go build -o ${DEST:-/build/bin}/${GOOS:-linux}/${GOARCH:-amd64}/dockercloud-agent-${VERSION:-latest}