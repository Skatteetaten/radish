package signaler

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

// Start : Used to forward signals to child processes
func Start(p *os.Process, forwardGracetime time.Duration) {
	c := make(chan os.Signal, 10)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGUSR1)
	go forward(c, p, forwardGracetime)
}
func forward(signals chan os.Signal, p *os.Process, forwardGracetime time.Duration) {
	for s := range signals {
		logrus.Infof("Got %s. Sending to child in %0.0f seconds", s, forwardGracetime.Seconds())
		if forwardGracetime > 0 {
			time.Sleep(forwardGracetime)
		}
		logrus.Infof("Sending signal %s to process %d", syscall.SIGTERM, p.Pid)
		err := p.Signal(s)
		if err != nil {
			logrus.Errorf("Error sending signal %s", err)
		}

	}
}
