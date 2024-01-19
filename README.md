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
# or (you can run `go tool dist list` to see all platforms and architectures)
# please note that not all tools will currently build on all platforms, currently Linux is prioritized
export GOOS=linux
export GOARCH=amd64
./build.sh
```

## Credits
Huge note on the `init` command:

I rewrote a lot of things but I yoinked the know-how and go-packages from James Mills's project `gonix`
that is under MIT license. Ive modified things here and there but I have to give this guy huge props
for a decently large chunk of code and for showing me how this works in Go.

I've also used the u-root implementation as reference for many commands. Outside of that, I've done my best to use all materials
as a tool for learning how something is done, and done my best to implement a version of my own to the best of my ability.
However it would have not been possible without the work of many many others.

Repositories used and referenced:
- ![gonix](https://git.mills.io/prologic/gonix.git) MIT License
- ![u-root](https://github.com/u-root/u-root) BSD 3-Clause "New" or "Revised" License
- ![mylanconnolly/parallel](https://github.com/mylanconnolly/parallel) - MIT License. referenced and used in my `parallel` implementation
- ![jesse-vdks go-flags](https://github.com/jesse-vdk/go-flags) - thank you for this package, I have a bone to pick with the Go stdlib
  implementation of `flag` jessevdks package is really nice to use and I have used it for nearly every tool.

Please note that none of these projects endorse or promote any of this work. I am greatly thankful to them
for all of their work and for making it open-source and allowing permissive usage of their work. As well as Go
for providing a very robust standard lib and great documentation.
