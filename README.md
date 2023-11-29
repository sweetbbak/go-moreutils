<p></p>
<p align="center">
  <img src="assets/core.png" />
</p>

Coreutils with a modern spin.
Why? Someone once said:
"If I cannot write it, then I do not understand it..."
and I took that to heart. This is purely for my entertainment and understanding of Unix systems,
if you want something battle tested, try: u-root, busybox, toybox, gnu coreutils or rust-coreutils.

## Examples
<p align="left">
  <img src="assets/img.png" />
</p>

`echo` support HEX colors with the `-f` switch
```bash
  ./echo -f '{#1e1e1e}HELLO{#clr}'
```

`printenv`has colored output
```bash
  ./printenv -p
```
![print env with colored output monokai theme](assets/printenv.png)

`sleep`
```sh
./sleep --exec 'notify-send "hello world" -i /path/to/image' 99s
# sleep accepts down to Nanoseconds and up to days
./sleep --exec 'notify-send "hello world" -i /path/to/image' 99us
# it also has a timer that prints to the terminal
./sleep --print --exec 'notify-send "hello world" -i /path/to/image' 9999ms
  
```

## Installation
```sh
git clone https://github.com/sweetbbak/go-moreutils.git
cd go-moreutils

# from there you can cd into any of the commands directories and run:
go build
# or alternatively:
go run *.go [args]
# TODO add the justfile with options to build any tool or all tools at once
just
```
