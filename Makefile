default: all

all: image
	mkdir -p ./build
	docker rm -f agentbuild || true
	docker run --name=agentbuild dockercloud-agent contrib/make-all.sh
	docker cp agentbuild:/build .
	docker rm -f agentbuild

clean:
	rm -fr build/
	docker rm -f agentbuild || true
	docker rmi dockercloud-agent || true

image:
	docker build --force-rm --rm -t dockercloud-agent .

test: image
	docker run --rm -t dockercloud-agent go test -v ./...

upload:
	s3cmd sync -P build s3://files.cloud.docker.com/dockercloud-agent/
