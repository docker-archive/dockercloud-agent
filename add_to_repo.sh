#!/bin/bash
set -e
# Production
# GPG_KEY_ID=A87A2270
# GPG_UID="Tutum Inc. (tutum) <info@tutum.co>"
# S3_BUCKET=repo.cloud.docker.com

# Staging
GPG_KEY_ID=90E64D7C
GPG_UID="Tutum Inc. (staging) <info@tutum.co>"
S3_BUCKET=repo-test.cloud.docker.com

if [ ! -f "$1" ]; then
	echo "Invalid package $1"
	exit 1
fi

cd repo/
rm -f *.rpm *.deb
cp ../$1 ./
gpg --send-keys --keyserver keyserver.ubuntu.com $GPG_KEY_ID
gpg --export -a $GPG_KEY_ID > ./gpg_public_key
gpg --export-secret-key -a $GPG_KEY_ID > ./gpg_private_key
docker build -t agentrepo .
docker run --rm -i -t -e AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID -e AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY -e S3_BUCKET=$S3_BUCKET -e GPG_UID="$GPG_UID" agentrepo
docker rmi agentrepo
rm -f *.rpm *.deb
