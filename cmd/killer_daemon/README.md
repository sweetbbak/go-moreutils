# Killer daemon
Block processes from running. killer_daemon runs in the background
and searches for processes that match the given parameters and then kills them.
You can match processes by name, substring matches and by giving partial command
line arguments to match against.

## Examples:
```sh
  ./killer_daemon --command "start-as=fullscreen"
  # in another terminal:
  kitty --start-as=fullscreen
  # kills the process immediately and outputs:
Found substring  match: kitty of [PID] 167520
2023/12/27 17:09:27 Killed process: kitty of [PID] 167520  
```

* blocking a program by name
```sh
  ./killer_daemon steam firefox
```

it will kill steam and firefox
