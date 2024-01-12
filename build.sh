#!/usr/bin/env bash
# build all

[ ! -d "build" ] && {
    mkdir build
}

exes=(ansi arch arp backoff base64 basename blkid \
cat chgrp chmod chown chroot cksum cp cpio cut daemonize \
date dd df diff dirname dmesg du echo env exit extras factor false \
file find free fusermount getty git_clone grep groups gunzip gzip \
head hexdump hostname httpd hwclock info kill killer_daemon \
less ln logname losetup lsdrivers lsmod lspci md5sum \
mkdir mkfifo mkfs mknod mktemp more mv netcat nice nl nohup nproc nsview \
parallel pgrep pkg printenv ps pwd readlink realpath reboot reset \
rm rmdir scanline scp seq setsid sh sha512sum showpath shred shuf \
shutdown sl sleep sprintf stat strace strings sudo \
sync tac tail tar tee time timeout touch tr \
tree true ts tty umount uname uniq unlink unshare \
unzip uptime uuidgen watch watchdog watchdogd wc wget which whoami \
whois xargs xxd yes zcat zip
)

for i in "${exes[@]}"; do
    [ -d "${i}" ] && {
        go build -o "build/${i}" "${i}"/*.go
    }
done
