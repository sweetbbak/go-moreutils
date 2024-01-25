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

## Stats
```
───────────────────────────────────────────────────────────────────────────────
Language                 Files     Lines   Blanks  Comments     Code Complexity
───────────────────────────────────────────────────────────────────────────────
Go                         396     50126     6587      3494    40045      10171
Plain Text                 202       496      202         0      294          0
Markdown                    17       387       61         0      326          0
Shell                        8       144       35        30       79         19
License                      2        40        7         0       33          0
gitignore                    2        21        0         0       21          0
───────────────────────────────────────────────────────────────────────────────
Total                      627     51214     6892      3524    40798      10190
───────────────────────────────────────────────────────────────────────────────
Estimated Cost to Develop (organic) $1,326,679
Estimated Schedule Effort (organic) 15.31 months
Estimated People Required (organic) 7.70
───────────────────────────────────────────────────────────────────────────────
Processed 1235749 bytes, 1.236 megabytes (SI)
───────────────────────────────────────────────────────────────────────────────
```

## Credits
thanks to James Mills's project `gonix` that is under MIT license for a good example of an `init`
command and some of the libs implemented in the init.

thanks to `u-root` for having one of the best implementations of Go linux user-land. Ive referenced
many of their tools and am currently using a few of their libs as well where re-writing them would
have been a monumental task (ie proper support for PCIE, block devices, network stacks etc...) at
some point I'd like to entirely re-write as much of these libs as I can.

Repositories used and referenced:
- ![gonix](https://git.mills.io/prologic/gonix.git) MIT License
- ![u-root](https://github.com/u-root/u-root) BSD 3-Clause "New" or "Revised" License
- ![mylanconnolly/parallel](https://github.com/mylanconnolly/parallel) - MIT License. referenced and used in my `parallel` implementation
- ![jesse-vdks go-flags](https://github.com/jesse-vdk/go-flags) - thank you for this package, I have a bone to pick with the Go stdlib
  implementation of `flag` jessevdks package is really nice to use and I have used it for nearly every tool.

Please note that none of these projects endorse or promote any of this work. I am greatly thankful to them
for all of their work and for making it open-source and allowing permissive usage of their work. As well as Go
for providing a very robust standard lib and great documentation.

