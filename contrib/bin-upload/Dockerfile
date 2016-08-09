FROM ubuntu:xenial

RUN apt-get update && apt-get install -yq curl awscli cpio rpm2cpio binutils xz-utils
ADD upload-docker-bin.sh /

ENV AWS_ACCESS_KEY_ID changeme
ENV AWS_SECRET_ACCESS_KEY changeme
ENV S3_BUCKET files.cloud.docker.com

CMD ["/upload-docker-bin.sh"]
