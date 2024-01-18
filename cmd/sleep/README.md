## Sleep

```sh
# default sleep unit is seconds, it can be implied:
./sleep 1
./sleep 0.5
./sleep 10.5
# milli
./sleep --exec "notify-send helloworld" 100000ms
# nano
./sleep --exec "notify-send helloworld" 999999ns
# micro
./sleep --exec "notify-send helloworld" 999999us
# seconds
./sleep --exec "notify-send helloworld" 33s &
# Print the timer as it runs
./sleep --print --exec "notify-send helloworld" 3m

# Loop X number of times
./sleep --loop 10 --exec "set_wallpaper.sh" 1h
# -1 is an infinite loop
./sleep --verbose --loop -1 --exec 'ls -lah' 0.1s
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
