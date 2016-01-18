FROM tutum/reprepro:latest

ENV AWS_ACCESS_KEY_ID changeme
ENV AWS_SECRET_ACCESS_KEY changeme
ENV S3_BUCKET repo.tutum.co

ADD distributions /repo/conf/distributions
ADD . /

CMD ["/run.sh"]
