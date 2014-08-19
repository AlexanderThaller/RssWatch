package misc

import (
	"os"
	"os/signal"
	"syscall"
)

// WaitForSigint waits for SIGINT or SIGTERM and returns.
func WaitForSigint() {
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)
	<-s
}

// ReplaceNth replaces the nth occurence of the given character in the
// given string. If nt is small or equal to zero we will just return the given string.
func ReplaceNth(st, ol, ne string, nt int) string {
	if nt <= 0 {
		return st
	}

	if ol == "" {
		return st
	}

	s := []rune(st)
	o := []rune(ol)
	n := []rune(ne)

	var j int

	for i, d := range st {
		if d == o[0] {
			j += 1

			if j == nt {
				if len(n) != 0 {
					s[i] = n[0]
				}

				j = 0
			}
		}
	}

	return string(s)
}
