#!/bin/sh

export REPO=repo-test.tutum.co.s3.amazonaws.com
export GPG_KEY_PACKAGE_ID=90E64D7C
export HOST="https://cloud-stage.docker.com"

if [ -z "$1" ]; then
	echo "token is not provided"
	exit 1
fi

if which sudo >/dev/null 2>&1; then
	#curl -Ls https://get.tutum.co/ | sudo -H sh -s $1
	cat install-agent.sh | sudo -H REPO=${REPO} GPG_KEY_PACKAGE_ID=${GPG_KEY_PACKAGE_ID} HOST=${HOST} sh -s $1
else
	cat install-agent.sh | REPO=${REPO} GPG_KEY_PACKAGE_ID=${GPG_KEY_PACKAGE_ID} HOST=${HOST} sh -s $1
fi
