#!/usr/bin/env bash
# build all

[ ! -d "build" ] && {
    mkdir build
}

exes=(
    arch
    basename
    chroot
    cmds
    df
    echo
    env
    false
    free
    git_clone
    grep
    groups
    hostname
    lsmod
    md5sum
    mkdir
    mv
    printenv
    pwd
    reboot
    rm
    shutdown
    sprintf
    switch_root
    sync
    true
    uname
    uptime
    whoami
    xargs
    yes
)

for i in "${exes[@]}"; do
    go build -o "build/${i}" "${i}"/*.go
done
