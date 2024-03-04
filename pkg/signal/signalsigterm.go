//go:build darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris
// +build darwin dragonfly freebsd linux netbsd openbsd solaris

package signal

import (
	"os"
	"syscall"
)

func init() {
	signals = []os.Signal{os.Interrupt, syscall.SIGTERM}
}
