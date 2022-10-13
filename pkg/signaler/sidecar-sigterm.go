package signaler

import (
	"github.com/mitchellh/go-ps"
	"github.com/sirupsen/logrus"
	"syscall"
)

// SendSigtermToSidecars :
func SendSigtermToSidecars(processNames []string) {
	if len(processNames) == 0 {
		logrus.Debugf("SIGTERM signaling ended, no processes specified for SIGTERM signaling")
		return
	}

	processes, err := ps.Processes()
	if err != nil {
		logrus.Errorf("Error getting process list %s", err)
	}

	for _, p := range processes {
		if shouldTerminateProcess(p.Executable(), processNames) {
			logrus.Infof("Sending SIGTERM to process=%s PID=%d", p.Executable(), p.Pid())
			err := syscall.Kill(p.Pid(), syscall.SIGTERM)
			if err != nil {
				logrus.Errorf("Error sending SIGTERM to process=%s, PID=%d", p.Executable(), p.Pid())
				return
			}
		} else {

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
