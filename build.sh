#!/usr/bin/env bash
# build all

bin="bin"

[ ! -d "$bin" ] && {
    mkdir build
}

shopt -s extglob
exes=( cmd/!(folder) )

for i in "${exes[@]}"; do
    [ -d "${i}" ] && {
        go build -o "$bin/${i##*/}" -ldflags="-s -w" "${i}"/*.go
    }
done
