package signaler

import (
	"github.com/mitchellh/go-ps"
	"github.com/sirupsen/logrus"
	"syscall"
	"time"
)

// SendSigtermToSidecars :
func SendSigtermToSidecars(processNames []string, terminateGracetime time.Duration) {
	if len(processNames) == 0 {
		logrus.Debugf("SIGTERM signaling ended, no processes specified for SIGTERM signaling")
		return
	}

	processes, err := ps.Processes()
	if err != nil {
		logrus.Errorf("Error getting process list %s", err)
	}

	logrus.Infof("Run completed. Sending SIGTERM to configured sidecar processes in %0.0f seconds", terminateGracetime.Seconds())
	if terminateGracetime > 0 {
		time.Sleep(terminateGracetime)
	}

	for _, p := range processes {
		if shouldTerminateProcess(p.Executable(), processNames) {
			logrus.Infof("Sending SIGTERM to process=%s PID=%d", p.Executable(), p.Pid())
			err := syscall.Kill(p.Pid(), syscall.SIGTERM)
			if err != nil {
				logrus.Errorf("Error sending SIGTERM to process=%s, PID=%d", p.Executable(), p.Pid())
				return
			}
		}
	}
}

func shouldTerminateProcess(processName string, processNames []string) bool {
	for _, p := range processNames {
		if processName == p {
			return true
		}
	}
	return false
}
