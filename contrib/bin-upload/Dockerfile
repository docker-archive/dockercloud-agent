FROM ubuntu:xenial

RUN apt-get update && apt-get install -yq curl awscli cpio rpm2cpio binutils xz-utils gnupg2 gnupg-agent && apt-get clean
ADD upload-docker-bin.sh /

ENV AWS_ACCESS_KEY_ID=changeme AWS_SECRET_ACCESS_KEY=changeme S3_BUCKET=files.cloud.docker.com GPG_UID=A87A2270

CMD ["/upload-docker-bin.sh"]
