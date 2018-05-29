package reaper

import (
	"github.com/Sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func Start() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGCHLD)
	go reapChilds(c)
}

func reapChilds(signals chan os.Signal) {
	for range signals {
		for {
			var status syscall.WaitStatus
			pid, err := syscall.Wait4(-1, &status, syscall.WNOHANG, nil)
			if err == nil {
				logrus.Infof("Reaped process %d with exit status %d", pid, status)
			} else {
				break
			}
		}
	}
}
