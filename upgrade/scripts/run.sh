#!/bin/sh

cp /scripts/upgrade-agent.sh /rootfs/tmp/upgrade-agent.sh 
chroot /rootfs /tmp/upgrade-agent.sh
