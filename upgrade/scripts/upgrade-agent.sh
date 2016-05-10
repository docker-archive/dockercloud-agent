#!/bin/bash

AGENT_PIDFILE="/var/run/dockercloud-agent.pid"
AGENT_NAME="dockercloud-agent"
AGENT_BINARY=$(which $AGENT_NAME)


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

get_agent_pid(){
    if [ -f ${AGENT_PIDFILE} ]; then
        cat $AGENT_PIDFILE
    else
        ps ax | grep -F $AGENT_NAME | grep -v grep | grep -v sudo | head -n 1 | cut -d " " -f 1 2>/dev/null
    fi
}


get_agent_version()
{
    ver=$(${AGENT_BINARY} -v 2>/dev/null)
    if [ $? -eq 0 ]; then
        echo ${ver}
    else
        echo "unknown version"
    fi
    unset ver
}

OLD_AGENT_VERSION=$(get_agent_version)
AGENT_PID=$(get_agent_pid)
if [ -n "${AGENT_PID}" ]; then
    echo "=> dockercloud-agent(${OLD_AGENT_VERSION}) is running with PID: ${AGENT_PID}"
else
    echo "=> dockercloud-agent(${OLD_AGENT_VERSION}) is running with PID: unknown"
fi

DISTRO="$(get_distribution_type)"
case ${DISTRO} in
    ubuntu|debian)
        echo "=> host operating system detected: ${DISTRO}"
        apt-get update || true
        apt-get install -y ${AGENT_NAME}
        ;;
    centos|rhel)
        echo "=> host operating system detected: ${DISTRO}"
        yum check-update
        yum install -y dockercloud-agent
        ;;
    fedora)
        echo "=> host operating system detected: ${DISTRO}"
        yum check-update
        yum install -y --best --allowerasing dockercloud-agent
        ;;
    *)
        echo "=> error: Cannot detect Linux distribution (${DISTRO}), or it's unsupported"
        exit 5
        ;;
esac

NEW_AGENT_VERSION=$(get_agent_version)
if [ "${OLD_AGENT_VERSION}" == "${NEW_AGENT_VERSION}" ]; then
    echo "=> version of dockercloud-agent remains the same"
    echo "=> exiting without any changes"
    exit 6
else
    echo "=> dockercloud-agent is upgraded from ${OLD_AGENT_VERSION} to ${NEW_AGENT_VERSION}"
    if [ -n "${AGENT_PID}" ]; then
        echo "=> killing the current dockercloud-agent process, and it will be restarted by upstart/systemd/sysmvinit"
        echo "=> NOTICE: you might have to restart your stopped containers if they are launched without autorestart option"
        kill ${AGENT_PID}
        exit 3
    else
        echo "=> Please restart dockercloud-agent to apply the changes"
        exit 4
    fi
fi
