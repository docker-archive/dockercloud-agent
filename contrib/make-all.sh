#!/bin/bash

mkdir -p /build/{bin,ubuntu,redhat}
contrib/make-bin.sh /build/bin

echo 'Building deb'
contrib/make-deb.sh /build/ubuntu

echo 'Building rpm'
contrib/make-rpm.sh /build/redhat
