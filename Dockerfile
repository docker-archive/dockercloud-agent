FROM resin/rpi-raspbian

RUN gpg --keyserver pgpkeys.mit.edu --recv-key  8B48AD6246925553      
RUN gpg -a --export 8B48AD6246925553 | sudo apt-key add -
RUN echo 'deb http://httpredir.debian.org/debian jessie-backports main' >> /etc/apt/sources.list

# Install FPM for packaging
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -qy build-essential ruby ruby-dev rpm && \
	gem install --no-rdoc --no-ri fpm --version 1.0.2
RUN apt-get install -qy -t jessie-backports golang

ENV GOPATH /go
WORKDIR /go/src/github.com/docker/dockercloud-agent
RUN apt-get install -qy git
ADD . /go/src/github.com/docker/dockercloud-agent
RUN go get -d -v && go build -v

CMD ["/go/src/github.com/docker/dockercloud-agent"]
