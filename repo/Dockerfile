FROM tutum/reprepro:latest

ENV AWS_ACCESS_KEY_ID changeme
ENV AWS_SECRET_ACCESS_KEY changeme
ENV S3_BUCKET repo.cloud.docker.com

ADD distributions /repo/conf/distributions
ADD . /

CMD ["/run.sh"]
