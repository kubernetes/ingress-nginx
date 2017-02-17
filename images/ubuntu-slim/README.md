
Small Ubuntu 16.04 docker image

The size of this image is ~44MB (less than half than `ubuntu:16.04).
This is possible by the removal of packages that are not required in a container:
- e2fslibs
- e2fsprogs
- init
- initscripts
- libcap2-bin
- libcryptsetup4
- libdevmapper1.02.1
- libkmod2
- libmount1
- libncursesw5
- libprocps4
- libsmartcols1
- libudev1
- mount
- ncurses-base
- ncurses-bin
- procps
- systemd
- systemd-sysv
- tzdata
- util-linux
