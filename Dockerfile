FROM golang:1.5.4

# Install FPM for packaging
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -qy ruby ruby-dev rpm && \
	gem install --no-rdoc --no-ri fpm --version 1.0.2

ENV GOPATH /go
WORKDIR /go/src/github.com/docker/dockercloud-agent
ADD . /go/src/github.com/docker/dockercloud-agent
RUN go get -d -v && go build -v

CMD ["/go/src/github.com/docker/dockercloud-agent"]
