package nginx

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

var settleTime = 1000 * time.Millisecond

func TestRotateSignal(t *testing.T) {

	file, err := os.CreateTemp("/tmp", "logrotate")
	if err != nil {
		log.Fatal(err)
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(file.Name())

	r := nginxLogRotate{
		paths:           []string{file.Name()},
		rotateAfterSize: 0,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)

	pid := syscall.Getpid()
	logrus.Infof("Pid: %d", pid)

	if err := r.rotate(pid, file.Name()); err != nil {
		t.Fatalf("Log rotate failed")
	}

	waitSig(t, c, syscall.SIGUSR1)
}

func TestHandleLogRotate(t *testing.T) {

	file, err := os.CreateTemp("/tmp", "logrotate")
	if err != nil {
		log.Fatal(err)
	}
	defer func(name string) {
		_ = os.Remove(name)
	}(file.Name())

	e := NewNginxExecutor(0, 600, []string{file.Name()})

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	pid := syscall.Getpid()
	e.StartLogRotate(pid)

	waitSig(t, c, syscall.SIGUSR1)
}

func TestNameArchive(t *testing.T) {

	path := "/u01/logs/nginx.access"

	extension := filepath.Ext(path)
	base := path[0 : len(path)-len(extension)]
	oldLog := fmt.Sprintf("%s.0%s", base, extension)

	assert.Equal(t, "/u01/logs/nginx.0.access", oldLog)
}

// https://golang.org/src/os/signal/signal_test.go
func waitSig(t *testing.T, c <-chan os.Signal, sig os.Signal) {

	t.Helper()

	waitSig1(t, c, sig, false)
}

// https://golang.org/src/os/signal/signal_test.go
func waitSig1(t *testing.T, c <-chan os.Signal, sig os.Signal, all bool) {

	t.Helper()

	// Sleep multiple times to give the kernel more tries to

	// deliver the signal.

	start := time.Now()

	timer := time.NewTimer(settleTime / 10)

	defer timer.Stop()

	// If the caller notified for all signals on c, filter out SIGURG,

	// which is used for runtime preemption and can come at unpredictable times.

	// General user code should filter out all unexpected signals instead of just

	// SIGURG, but since os/signal is tightly coupled to the runtime it seems

	// appropriate to be stricter here.

	for time.Since(start) < settleTime {

		select {

		case s := <-c:

			if s == sig {

				return

			}

			if !all || s != syscall.SIGURG {

				t.Fatalf("signal was %v, want %v", s, sig)

			}

		case <-timer.C:

			timer.Reset(settleTime / 10)

		}

	}

	t.Fatalf("timeout after %v waiting for %v", settleTime, sig)

}
