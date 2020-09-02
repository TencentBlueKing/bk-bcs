package util

import "syscall"

func SendTermSignal() {
	syscall.Kill(syscall.Getpid(), syscall.SIGUSR1)
}
