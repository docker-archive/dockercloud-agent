#!/bin/sh
#
# Usage:
# curl -Ls https://get.cloud.docker.com/ | sudo -H sh -s [Token] [UUID] [CertCommonName]
#
set -e
GPG_KEY_ID=A87A2270
GPG_KEY_PACKAGE_ID=${GPG_KEY_PACKAGE_ID:-A87A2270}
GPG_KEY_URL=https://files.cloud.docker.com/keys/$GPG_KEY_ID.pub
GPG_KEY_PACKAGE_URL=https://files.cloud.docker.com/keys/$GPG_KEY_PACKAGE_ID.pub
REPO=${REPO:-repo.cloud.docker.com}
HOST=${HOST:-https://cloud.docker.com/}
SUPPORT_URL=https://docs.docker.com/docker-cloud/tutorials/byoh/
export DEBIAN_FRONTEND=noninteractive


if [ -f "/etc/dockercloud/agent/dockercloud-agent.conf" ]; then
	cat <<EOF
ERROR: Docker Cloud Agent is already installed in this host
If the node failed to register properly with Docker Cloud, try to restart the agent by executing:

	service dockercloud-agent restart

and check the logs at /var/log/dockercloud/agent.log for possible errors.
If the problem persists, please contact us at support@docker.com
EOF
	exit 1
fi

if [ "$(uname -m)" != "x86_64" ]; then
	cat <<EOF
ERROR: Unsupported architecture: $(uname -m)
Only x86_64 architectures are supported at this time
Learn more: $SUPPORT_URL
EOF
	exit 1
fi

get_distribution_type()
{
	local lsb_dist
	lsb_dist="$(lsb_release -si 2> /dev/null || echo "unknown")"
	if [ "$lsb_dist" = "unknown" ]; then
		if [ -r /etc/lsb-release ]; then
			lsb_dist="$(. /etc/lsb-release && echo "$DISTRIB_ID")"
		elif [ -r /etc/debian_version ]; then
			lsb_dist='debian'
		elif [ -r /etc/fedora-release ]; then
			lsb_dist='fedora'
		elif [ -r /etc/centos-release ]; then
			lsb_dist='centos'
		elif [ -r /etc/redhat-release ]; then
			lsb_dist='rhel'
		elif [ -r /etc/os-release ]; then
			lsb_dist="$(. /etc/os-release && echo "$ID")"
		fi
	fi
	lsb_dist="$(echo "$lsb_dist" | tr '[:upper:]' '[:lower:]')"
	echo $lsb_dist
}

case "$(get_distribution_type)" in
	ubuntu|debian)
		echo "-> Adding Docker Cloud's GPG key..."
		curl -Ls --retry 30 --retry-delay 10 $GPG_KEY_URL | gpg --import -
		curl -Ls --retry 30 --retry-delay 10 $GPG_KEY_PACKAGE_URL | apt-key add -
		echo "-> Installing required dependencies..."
		modprobe -q aufs || (apt-get update -qq && apt-get install -yq linux-image-extra-$(uname -r) || \
			echo "!! Failed to install linux-image-extra package. AUFS support (which is recommended) may not be available.")
		echo "-> Installing dockercloud-agent..."
		echo deb [arch=amd64] http://$REPO/ubuntu/ dockercloud main > /etc/apt/sources.list.d/dockercloud.list
		apt-get update -qq && apt-get install -yq dockercloud-agent
		;;
	fedora|centos|rhel)
		echo "-> Adding Docker Cloud's GPG key..."
		yum install -y gpg rpm curl
		curl -Ls --retry 30 --retry-delay 10 $GPG_KEY_URL | gpg --import -
		rpm --import $GPG_KEY_PACKAGE_URL
		echo "-> Installing dockercloud-agent..."
		cat > /etc/yum.repos.d/dockercloud.repo <<EOF
[dockercloud]
name=dockercloud
baseurl=http://$REPO/redhat/\$basearch
enabled=1
gpgkey=$GPG_KEY_PACKAGE_URL
repo_gpgcheck=1
gpgcheck=1
EOF
		yum install -y dockercloud-agent
		;;
	*)
		echo "ERROR: Cannot detect Linux distribution or it's unsupported"
		echo "Learn more: $SUPPORT_URL"
		exit 1
		;;
esac

echo "-> Configuring dockercloud-agent..."
mkdir -p /etc/dockercloud/agent
cat > /etc/dockercloud/agent/dockercloud-agent.conf <<EOF
{
	"Host":"${HOST}",
	"Token":"${1}",
	"UUID":"${2}",
	"CertCommonName":"${3}"
}
EOF

if [ -d /run/systemd/system ] ; then
	echo "-> Enabling dockercloud-agent to start on boot on systemd..."
	systemctl enable dockercloud-agent.service || true
fi

if [ ! -z "${1}" ]; then
	echo "-> Starting dockercloud-agent service..."
	service dockercloud-agent stop > /dev/null 2>&1 || true
	service dockercloud-agent start
fi

echo "-> Done!"
cat <<EOF

*******************************************************************************
Docker Cloud Agent installed successfully
*******************************************************************************

You can now deploy containers to this node using Docker Cloud

EOF
