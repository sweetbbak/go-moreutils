// thanks u-root
package main

import (
	"fmt"
	"os"
	"syscall"
)

var (
	signames = []string{
		"SIGHUP",
		"SIGINT",
		"SIGQUIT",
		"SIGILL",
		"SIGTRAP",
		"SIGABRT",
		"SIGBUS",
		"SIGFPE",
		"SIGKILL",
		"SIGUSR1",
		"SIGSEGV",
		"SIGUSR2",
		"SIGPIPE",
		"SIGALRM",
		"SIGTERM",
		"SIGSTKFLT",
		"SIGCHLD",
		"SIGCONT",
		"SIGSTOP",
		"SIGTSTP",
		"SIGTTIN",
		"SIGTTOU",
		"SIGURG",
		"SIGXCPU",
		"SIGXFSZ",
		"SIGVTALRM",
		"SIGPROF",
		"SIGWINCH",
		"SIGIO",
		"SIGPWR",
		"SIGSYS",
		"SIGRTMIN",
		"SIGRTMIN+1",
		"SIGRTMIN+2",
		"SIGRTMIN+3",
		"SIGRTMIN+4",
		"SIGRTMIN+5",
		"SIGRTMIN+6",
		"SIGRTMIN+7",
		"SIGRTMIN+8",
		"SIGRTMIN+9",
		"SIGRTMIN+10",
		"SIGRTMIN+11",
		"SIGRTMIN+12",
		"SIGRTMIN+13",
		"SIGRTMIN+14",
		"SIGRTMIN+15",
		"SIGRTMAX-14",
		"SIGRTMAX-13",
		"SIGRTMAX-12",
		"SIGRTMAX-11",
		"SIGRTMAX-10",
		"SIGRTMAX-9",
		"SIGRTMAX-8",
		"SIGRTMAX-7",
		"SIGRTMAX-6",
		"SIGRTMAX-5",
		"SIGRTMAX-4",
		"SIGRTMAX-3",
		"SIGRTMAX-2",
		"SIGRTMAX-1",
		"SIGRTMAX",
	}
	signums = map[string]os.Signal{
		"SIGHUP":      syscall.Signal(1),
		"SIGINT":      syscall.Signal(2),
		"SIGQUIT":     syscall.Signal(3),
		"SIGILL":      syscall.Signal(4),
		"SIGTRAP":     syscall.Signal(5),
		"SIGABRT":     syscall.Signal(6),
		"SIGBUS":      syscall.Signal(7),
		"SIGFPE":      syscall.Signal(8),
		"SIGKILL":     syscall.Signal(9),
		"SIGUSR1":     syscall.Signal(10),
		"SIGSEGV":     syscall.Signal(11),
		"SIGUSR2":     syscall.Signal(12),
		"SIGPIPE":     syscall.Signal(13),
		"SIGALRM":     syscall.Signal(14),
		"SIGTERM":     syscall.Signal(15),
		"SIGSTKFLT":   syscall.Signal(16),
		"SIGCHLD":     syscall.Signal(17),
		"SIGCONT":     syscall.Signal(18),
		"SIGSTOP":     syscall.Signal(19),
		"SIGTSTP":     syscall.Signal(20),
		"SIGTTIN":     syscall.Signal(21),
		"SIGTTOU":     syscall.Signal(22),
		"SIGURG":      syscall.Signal(23),
		"SIGXCPU":     syscall.Signal(24),
		"SIGXFSZ":     syscall.Signal(25),
		"SIGVTALRM":   syscall.Signal(26),
		"SIGPROF":     syscall.Signal(27),
		"SIGWINCH":    syscall.Signal(28),
		"SIGIO":       syscall.Signal(29),
		"SIGPWR":      syscall.Signal(30),
		"SIGSYS":      syscall.Signal(31),
		"SIGRTMIN":    syscall.Signal(34),
		"SIGRTMIN+1":  syscall.Signal(35),
		"SIGRTMIN+2":  syscall.Signal(36),
		"SIGRTMIN+3":  syscall.Signal(37),
		"SIGRTMIN+4":  syscall.Signal(38),
		"SIGRTMIN+5":  syscall.Signal(39),
		"SIGRTMIN+6":  syscall.Signal(40),
		"SIGRTMIN+7":  syscall.Signal(41),
		"SIGRTMIN+8":  syscall.Signal(42),
		"SIGRTMIN+9":  syscall.Signal(43),
		"SIGRTMIN+10": syscall.Signal(44),
		"SIGRTMIN+11": syscall.Signal(45),
		"SIGRTMIN+12": syscall.Signal(46),
		"SIGRTMIN+13": syscall.Signal(47),
		"SIGRTMIN+14": syscall.Signal(48),
		"SIGRTMIN+15": syscall.Signal(49),
		"SIGRTMAX-14": syscall.Signal(50),
		"SIGRTMAX-13": syscall.Signal(51),
		"SIGRTMAX-12": syscall.Signal(52),
		"SIGRTMAX-11": syscall.Signal(53),
		"SIGRTMAX-10": syscall.Signal(54),
		"SIGRTMAX-9":  syscall.Signal(55),
		"SIGRTMAX-8":  syscall.Signal(56),
		"SIGRTMAX-7":  syscall.Signal(57),
		"SIGRTMAX-6":  syscall.Signal(58),
		"SIGRTMAX-5":  syscall.Signal(59),
		"SIGRTMAX-4":  syscall.Signal(60),
		"SIGRTMAX-3":  syscall.Signal(61),
		"SIGRTMAX-2":  syscall.Signal(62),
		"SIGRTMAX-1":  syscall.Signal(63),
		"SIGRTMAX":    syscall.Signal(64),
		"1":           syscall.Signal(1),
		"2":           syscall.Signal(2),
		"3":           syscall.Signal(3),
		"4":           syscall.Signal(4),
		"5":           syscall.Signal(5),
		"6":           syscall.Signal(6),
		"7":           syscall.Signal(7),
		"8":           syscall.Signal(8),
		"9":           syscall.Signal(9),
		"10":          syscall.Signal(10),
		"11":          syscall.Signal(11),
		"12":          syscall.Signal(12),
		"13":          syscall.Signal(13),
		"14":          syscall.Signal(14),
		"15":          syscall.Signal(15),
		"16":          syscall.Signal(16),
		"17":          syscall.Signal(17),
		"18":          syscall.Signal(18),
		"19":          syscall.Signal(19),
		"20":          syscall.Signal(20),
		"21":          syscall.Signal(21),
		"22":          syscall.Signal(22),
		"23":          syscall.Signal(23),
		"24":          syscall.Signal(24),
		"25":          syscall.Signal(25),
		"26":          syscall.Signal(26),
		"27":          syscall.Signal(27),
		"28":          syscall.Signal(28),
		"29":          syscall.Signal(29),
		"30":          syscall.Signal(30),
		"31":          syscall.Signal(31),
		"34":          syscall.Signal(34),
		"35":          syscall.Signal(35),
		"36":          syscall.Signal(36),
		"37":          syscall.Signal(37),
		"38":          syscall.Signal(38),
		"39":          syscall.Signal(39),
		"40":          syscall.Signal(40),
		"41":          syscall.Signal(41),
		"42":          syscall.Signal(42),
		"43":          syscall.Signal(43),
		"44":          syscall.Signal(44),
		"45":          syscall.Signal(45),
		"46":          syscall.Signal(46),
		"47":          syscall.Signal(47),
		"48":          syscall.Signal(48),
		"49":          syscall.Signal(49),
		"50":          syscall.Signal(50),
		"51":          syscall.Signal(51),
		"52":          syscall.Signal(52),
		"53":          syscall.Signal(53),
		"54":          syscall.Signal(54),
		"55":          syscall.Signal(55),
		"56":          syscall.Signal(56),
		"57":          syscall.Signal(57),
		"58":          syscall.Signal(58),
		"59":          syscall.Signal(59),
		"60":          syscall.Signal(60),
		"61":          syscall.Signal(61),
		"62":          syscall.Signal(62),
		"63":          syscall.Signal(63),
		"64":          syscall.Signal(64),
	}
)

func siglist() (s string) {
	for i, sig := range signames {
		s = s + fmt.Sprintf("%d: %v\n", i, sig)
	}
	return
}
