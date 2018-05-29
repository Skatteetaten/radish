package signaler

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
)

//Start : Used to forward signals to child processes
func Start(p *os.Process) {
	c := make(chan os.Signal, 10)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	go forward(c, p)
}
func forward(signals chan os.Signal, p *os.Process) {
	for s := range signals {
		logrus.Infof("Got %s", s)
		logrus.Infof("Sending signal %s to process %d", syscall.SIGTERM, p.Pid)
		err := p.Signal(s)
		if err != nil {
			logrus.Errorf("Error sending signal %s", err)
		}

	}
}
