#!/usr/bin/env bash
set -ex

DOCKER_VERSIONS="1.11.1-cs1 1.11.2-cs4"


function process_tgz() {
    sha256sum ${1} > ${1}.sha256
    md5sum ${1} > ${1}.md5
    gpg2 -s -b -u ${GPG_UID} ${1}
    mkdir build
    mv ${1}* build/
    cd build
    ls -la
    aws s3 sync ./ s3://$S3_BUCKET${2} --acl public-read --region us-east-1
}

function process_rpm() {
    rpm2cpio ./${1} | cpio -idmv
    cd ./usr
    mv bin docker
    tar czvf docker-${2}.tgz docker
    process_tgz docker-${2}.tgz ${3}
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
    process_tgz docker-${2}.tgz ${3}
}

for version in ${DOCKER_VERSIONS}; do
    # Static binaries
    cd $(mktemp -d)
    TGZ_NAME=docker-${version}.tgz
    curl -O https://s3.amazonaws.com/packages.docker.com/${version:0:4}/builds/linux/amd64/${TGZ_NAME}
    process_tgz ${TGZ_NAME} /packages/docker/

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

        case "${ubuntu_version}" in
        precise) ubuntu_version_num=12.04
                ;;
        trusty) ubuntu_version_num=14.04
                ;;
        xenial) ubuntu_version_num=16.04
                ;;
        *) exit 1
           ;;
        esac
        process_deb ${DEB_NAME} ${version} /packages/docker/ubuntu/${ubuntu_version_num}/
    done

    # Debian
    for debian_version in jessie wheezy; do
        cd $(mktemp -d)
        DEB_NAME=docker-engine_$(echo ${version} | tr '-' '~')-0~${debian_version}_amd64.deb
        curl -O https://s3.amazonaws.com/packages.docker.com/${version:0:4}/apt/repo/pool/main/d/docker-engine/${DEB_NAME}

        case "${debian_version}" in
        wheezy) debian_version_num=7
                ;;
        jessie) debian_version_num=8
                ;;
        *) exit 1
           ;;
        esac
        process_deb ${DEB_NAME} ${version} /packages/docker/debian/${debian_version_num}/
    done
done
