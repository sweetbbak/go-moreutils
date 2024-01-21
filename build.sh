#!/usr/bin/env bash
# build all coreutils in the "cmd" directory and output the binaries into "$bin"
# set -euo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
# sanity check
[ "$SCRIPT_DIR" = "" ] && echo "unable to find the root of the project" && exit 3
cd "$SCRIPT_DIR" || exit 3

#################
# User Variables
#################

# default output bin name
# use `go tool dist list` to see all platform and architecture options (*Note* that some tools will currently not build outside of linux)
bin="$SCRIPT_DIR/bin"
GOOS=linux
GOARCH=amd64

export GOOS
export GOARCH

# if it doesn't exist, we make it
[ ! -d "$bin" ] && mkdir "$bin"

# extended globbing so we can expand all the 'folders' in the 'cmd' directory
shopt -s extglob || exit 1
exes=( cmd/!(folder) )
shopt -u extglob

# loop over the cmd directories, check if they exist and build all of the go files
for i in "${exes[@]}"; do
    printf "\x1b[33m%s\x1b[0m %s\n" "building:" "$bin/${i##*/}"
    [ -d "${i}" ] && {
        if [ ! "$i" = "cmd/extras" ]; then
            go build -o "$bin/${i##*/}" -ldflags="-s -w" "${i}"/*.go
            [ -e "$bin/${i##*/}" ] && printf "\x1b[4;32m%s\x1b[0m %s\n" "successfully built:" "$bin/${i##*/}"
        fi
    }
done
