
Small Ubuntu 16.04 docker image

The size of this image is ~56MB (less than half than `ubuntu:16.04).
This is possible by the removal of packages that are not required in a container:
- dmsetup
- e2fsprogs
- init
- initscripts
- libcap2-bin
- libcryptsetup4
- libdevmapper1.02.1
- libkmod2
- libsmartcols1
- libudev1
- mount
- procps
- systemd
- systemd-sysv
- tzdata
- udev
- util-linux
