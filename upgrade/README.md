Usage
=====

    docker run -d -v /:/rootfs --pid=host --privileged dockercloud/agent:upgrade


Exit Code
---------

* 0 - agent upgraded (143, 137)
* 4 - manually restart required
* 5 - unknown/upsupported distro
* 6 - latest version already
