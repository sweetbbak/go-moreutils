#!/usr/bin/env bash
# build all

[ ! -d "build" ] && {
    mkdir build
}

exes=(
    ansi
    arch
    backoff
    base64
    basename
    blkid
    cat
    chgrp
    chmod
    chown
    chroot
    cp
    daemonize
    df
    dirname
    dmesg
    du
    echo
    env
    exit
    false
    find
    free
    fusermount
    getty
    git_clone
    grep
    groups
    head
    hexdump
    hostname
    hwclock
    # init
    kill
    killer_daemon
    less
    ln
    losetup
    # ls
    # ls2
    # ls3
    lsdrivers
    lsiso
    lsmod
    lspci
    md5sum
    mkdir
    mkfifo
    mkfs
    mknod
    mktemp
    more
    mv
    netcat
    nohup
    nproc
    nsview
    pgrep
    pkg
    printenv
    ps
    pwd
    readlink
    realpath
    reboot
    rm
    rmdir
    scp
    setsid
    sh
    sha512sum
    showpath
    shutdown
    sleep
    sprintf
    stat
    strace
    stty
    sudo
    switch_root
    sync
    tac
    tar
    tee
    time
    timeout
    true
    ts
    tty
    umount
    uname
    uniq
    unlink
    unshare
    uptime
    uuidgen
    watch
    wc
    which
    whoami
    whois
    xargs
    xxd
    yes
)

for i in "${exes[@]}"; do
    [ -d "${i}" ] && {
        go build -o "build/${i}" "${i}"/*.go
    }
done
