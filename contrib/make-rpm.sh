#!/bin/bash
set -e

DEST=$1

VERSION=$(cat VERSION)
PKGVERSION="${VERSION:-latest}"
PACKAGE_ARCHITECTURE="${ARCHITECTURE:-amd64}"
PACKAGE_URL="http://www.docker.com/"
PACKAGE_MAINTAINER="support@docker.com"
PACKAGE_DESCRIPTION="Agent to manage Docker hosts through Docker Cloud"
PACKAGE_LICENSE="Proprietary"

bundle_redhat() {
  DIR=$DEST/staging

  # Include our init scripts
  mkdir -p $DIR/etc/init $DIR/etc/init.d $DIR/lib/systemd/system/
  cp contrib/init/upstart/dockercloud-agent.conf $DIR/etc/init/
  cp contrib/init/sysvinit-redhat/dockercloud-agent $DIR/etc/init.d/
  cp contrib/init/systemd/dockercloud-agent.socket $DIR/lib/systemd/system/
  cp contrib/init/systemd/dockercloud-agent.service $DIR/lib/systemd/system/

  # Include logrotate
  mkdir -p $DIR/etc/logrotate.d/
  cp contrib/logrotate/dockercloud-agent $DIR/etc/logrotate.d/

  # Copy the binary
  # This will fail if the binary bundle hasn't been built
  mkdir -p $DIR/usr/bin
  cp /build/bin/linux/$PACKAGE_ARCHITECTURE/dockercloud-agent-$PKGVERSION $DIR/usr/bin/dockercloud-agent

  cat > $DEST/postinst <<'EOF'
#!/bin/sh
set -e

DOCKER_UPSTART_CONF="/etc/init/docker.conf"
if [ -f "${DOCKER_UPSTART_CONF}" ]; then
  echo "Removing conflicting docker upstart configuration file at ${DOCKER_UPSTART_CONF}..."
  rm -f ${DOCKER_UPSTART_CONF}
fi

if ! getent group docker > /dev/null; then
  groupadd --system docker
fi

if [ -n "$2" ]; then
  service dockercloud-agent restart 2>/dev/null || true
fi
service dockercloud-agent $_dh_action 2>/dev/null || true

#DEBHELPER#
EOF

  cat > $DEST/prerm <<'EOF'
#!/bin/sh
set -e

case "$1" in
  remove)
    service dockercloud-agent stop 2>/dev/null || true
  ;;
esac

#DEBHELPER#
EOF

  cat > $DEST/postrm <<'EOF'
#!/bin/sh
set -e

case "$1" in
  remove)
    rm -fr /usr/bin/docker /usr/lib/dockercloud
  ;;
  purge)
    rm -fr /usr/bin/docker /usr/lib/dockercloud /etc/dockercloud
  ;;
esac

# In case this system is running systemd, we make systemd reload the unit files
# to pick up changes.
if [ -d /run/systemd/system ] ; then
  systemctl --system daemon-reload > /dev/null || true
fi

#DEBHELPER#
EOF

  chmod +x $DEST/postinst $DEST/prerm $DEST/postrm

  (
    # switch directories so we create *.deb in the right folder
    cd $DEST

    # create dockercloud-agent-$PKGVERSION package
    fpm -s dir -C $DIR \
      --name dockercloud-agent --version $PKGVERSION \
      --after-install $DEST/postinst \
      --before-remove $DEST/prerm \
      --after-remove $DEST/postrm \
      --architecture "$PACKAGE_ARCHITECTURE" \
      --prefix / \
      --description "$PACKAGE_DESCRIPTION" \
      --maintainer "$PACKAGE_MAINTAINER" \
      --conflicts docker \
      --conflicts docker.io \
      --conflicts lxc-docker \
      --conflicts docker-engine \
      --depends "/bin/sh" \
      --depends iptables \
      --depends gnupg \
      --depends libc.so.6 \
      --depends libcgroup \
      --depends libpthread.so.0 \
      --depends libsqlite3.so.0 \
      --depends "libltdl.so.7()(64bit)" \
      --depends "libseccomp.so.2()(64bit)" \
      --depends tar \
      --depends xz \
      --depends "device-mapper-libs >= 1.02.90-1" \
      --provides dockercloud-agent \
      --replaces dockercloud-agent \
      --url "$PACKAGE_URL" \
      --license "$PACKAGE_LICENSE" \
      --config-files "etc/init/dockercloud-agent.conf" \
      --config-files "etc/init.d/dockercloud-agent" \
      --config-files "lib/systemd/system/dockercloud-agent.socket" \
      --config-files "lib/systemd/system/dockercloud-agent.service" \
      -t rpm .
  )

  rm $DEST/postinst $DEST/prerm $DEST/postrm
  rm -r $DIR
}

bundle_redhat
