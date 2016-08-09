#!/usr/bin/env bash
set -ex

DOCKER_VERSIONS=1.11.1-cs1


function process_rpm() {
    rpm2cpio ./${1} | cpio -idmv
    cd ./usr
    mv bin docker
    tar czvf docker-${2}.tgz docker
    sha256sum docker-${2}.tgz > docker-${2}.tgz.sha256
    md5sum docker-${2}.tgz > docker-${2}.tgz.md5
    mkdir build
    mv docker-${2}.tgz* build/
    cd build
    ls -la
    aws s3 sync ./ s3://$S3_BUCKET${3} --acl public-read --region us-east-1
}

function process_deb() {
    ar vx ./${1}
    if [ -f "data.tar.gz" ]; then
        tar zvxf data.tar.gz
    elif [ -f "data.tar.xz" ]; then
        tar Jvxf data.tar.xz
    else
        echo "data.tar not found"
        ls
        exit 1
    fi
    cd ./usr
    mv bin docker
    tar czvf docker-${2}.tgz docker
    sha256sum docker-${2}.tgz > docker-${2}.tgz.sha256
    md5sum docker-${2}.tgz > docker-${2}.tgz.md5
    mkdir build
    mv docker-${2}.tgz* build/
    cd build
    ls -la
    aws s3 sync ./ s3://$S3_BUCKET${3} --acl public-read --region us-east-1
}

for version in ${DOCKER_VERSIONS}; do
    # CentOS 7
    for centos_version in 7; do
        cd $(mktemp -d)
        RPM_NAME=docker-engine-$(echo ${version} | tr '-' '.')-1.el${centos_version}.centos.x86_64.rpm
        curl -O https://s3.amazonaws.com/packages.docker.com/${version:0:4}/yum/repo/main/centos/${centos_version}/Packages/${RPM_NAME}
        process_rpm ${RPM_NAME} ${version} /packages/docker/centos/${centos_version}/
    done

    # Fedora
    for fedora_version in 21 22 23; do
        cd $(mktemp -d)
        RPM_NAME=docker-engine-$(echo ${version} | tr '-' '.')-1.fc${fedora_version}.x86_64.rpm
        curl -O https://s3.amazonaws.com/packages.docker.com/${version:0:4}/yum/repo/main/fedora/${fedora_version}/Packages/${RPM_NAME}
        process_rpm ${RPM_NAME} ${version} /packages/docker/fedora/${fedora_version}/
    done

    # OpenSUSE
    for opensuse_version in 12.3 13.2; do
        cd $(mktemp -d)
        RPM_NAME=docker-engine-$(echo ${version} | tr '-' '.')-1.x86_64.rpm
        curl -O https://s3.amazonaws.com/packages.docker.com/${version:0:4}/yum/repo/main/opensuse/${opensuse_version}/Packages/${RPM_NAME}
        process_rpm ${RPM_NAME} ${version} /packages/docker/opensuse/${opensuse_version}/
    done

    # Oracle Linux
    for oraclelinux_version in 6 7; do
        cd $(mktemp -d)
        RPM_NAME=docker-engine-$(echo ${version} | tr '-' '.')-1.el${oraclelinux_version}.x86_64.rpm
        curl -O https://s3.amazonaws.com/packages.docker.com/${version:0:4}/yum/repo/main/oraclelinux/${oraclelinux_version}/Packages/${RPM_NAME}
        process_rpm ${RPM_NAME} ${version} /packages/docker/oraclelinux/${oraclelinux_version}/
    done

    # Ubuntu
    for ubuntu_version in precise trusty xenial; do
        cd $(mktemp -d)
        DEB_NAME=docker-engine_$(echo ${version} | tr '-' '~')-0~${ubuntu_version}_amd64.deb
        curl -O https://s3.amazonaws.com/packages.docker.com/${version:0:4}/apt/repo/pool/main/d/docker-engine/${DEB_NAME}
        process_deb ${DEB_NAME} ${version} /packages/docker/ubuntu/${ubuntu_version}/
    done

    # Debian
    for debian_version in jessie wheezy; do
        cd $(mktemp -d)
        DEB_NAME=docker-engine_$(echo ${version} | tr '-' '~')-0~${debian_version}_amd64.deb
        curl -O https://s3.amazonaws.com/packages.docker.com/${version:0:4}/apt/repo/pool/main/d/docker-engine/${DEB_NAME}
        process_deb ${DEB_NAME} ${version} /packages/docker/debian/${debian_version}/
    done
done
