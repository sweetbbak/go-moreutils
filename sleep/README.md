## Sleep

```sh
# milli
./sleep --exec "notify-send helloworld" 100000ms
# nano
./sleep --exec "notify-send helloworld" 999999ns
# micro
./sleep --exec "notify-send helloworld" 999999us
# seconds
./sleep --exec "notify-send helloworld" 33s &
# Print the timer as it runs
./sleep --exec "notify-send helloworld" 3m --print
```

### Duration
- examples
* 300ms
* 1.5h
* 2h45m

- valid time units:
  - ns
  - us (or μs)
  - ms
  - s
  - m
  - h

### TODO
- add units that amount to days

### Notes
`sleep` uses `sh -c` to run commands, this makes it easier to run commands in a pipeline and is more forgiving with bad syntax.
Im thinking of removing that, or providing an option to do a direct syscall.Exec or fork in rare cases that shell may not exist
like in a container or some kind of rescue situation.
