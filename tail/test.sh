#!/usr/bin/env bash
# test tail --follow

cleanup() {
    [ -p test-fifo ] && rm test-fifo
}

trap cleanup SIGINT EXIT

[ ! -p test-fifo ] && {
    mkfifo test-fifo
}

spamtext() {
    i=0; while true; do echo "loop $(random_emoji) ($i)" > test-fifo; i=$((i+1)); if [[ i -eq 20000000 ]]; then break; fi; done
}

export -f random_emoji

random_emoji() {
    rand=$(shuf -i 2600-2700 -n 1)
    echo -en "\u$rand"
}

spamtext &
./tail --follow test-fifo
