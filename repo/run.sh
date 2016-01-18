#!/bin/bash
set -e

gpg --import /gpg_public_key
gpg --import /gpg_private_key
if [ ! -z "$GPG_UID" ]; then
	echo "%_gpg_name $GPG_UID" >> /root/.rpmmacros
fi

if ls /*.deb 1> /dev/null 2>&1; then
	echo "=> package.deb found"
	mkdir -p /repo/db/
	mkdir -p /repo/dists/
	aws s3 sync s3://$S3_BUCKET/ubuntu/db/ /repo/db/ --region us-east-1
	aws s3 sync s3://$S3_BUCKET/ubuntu/dists/ /repo/dists/ --region us-east-1
	reprepro --keepunusednewfiles --ask-passphrase -Vb /repo includedeb tutum /*.deb
	aws s3 sync /repo/ s3://$S3_BUCKET/ubuntu/ --acl public-read --region us-east-1
fi

if ls /*.rpm 1> /dev/null 2>&1; then
	echo "=> package.rpm found"
	mkdir -p /repo/
	rpm --addsign /*.rpm
	cp /*.rpm /repo
	createrepo /repo
	gpg --detach-sign --armor /repo/repodata/repomd.xml
	aws s3 sync /repo/ s3://$S3_BUCKET/redhat/x86_64/ --acl public-read --region us-east-1
fi
