#!/bin/bash

VERSION=$(cat VERSION)
DEST=$1
GOOS=${GOOS:-linux} GOARCH=${GOARCH:-arm} go build -o ${DEST:-/build/bin}/${GOOS:-linux}/${GOARCH:-arm}/dockercloud-agent-${VERSION:-latest}
