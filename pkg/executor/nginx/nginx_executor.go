package nginx

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"syscall"
	"time"
)

//Executor :
type Executor interface {
	PrepareForNginxRun() *exec.Cmd
	HandleExit(exitCode int, pid int) int
	StartLogRotate(pid int, timeInMs int64)
}

type nginxExecutor struct {
	nginxExitHandler
	nginxLogRotate
}

type nginxExitHandler struct {
}

type nginxLogRotate struct {
	paths           []string
	rotateAfterSize int64
}

//NewNginxExecutor :
func NewNginxExecutor(rotateAfterSize int64, logfiles []string) Executor {
	return nginxExecutor{
		nginxExitHandler{},
		nginxLogRotate{
			paths:           logfiles,
			rotateAfterSize: rotateAfterSize,
		},
	}
}

func (m nginxExecutor) PrepareForNginxRun() *exec.Cmd {
	cmd := exec.Command("sh", "-c", "exec nginx -g 'daemon off;' -c /tmp/nginx/nginx.conf")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func (m nginxLogRotate) StartLogRotate(pid int, checkRotateAfterMs int64) {
	ticker := time.NewTicker(time.Duration(checkRotateAfterMs) * time.Millisecond)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				//Check file size of all files
				for _, path := range m.paths {
					file, err := os.Open(path)
					if err != nil {
						logrus.Errorf("Could not open log file: %v", err)
						continue
					}
					fileinfo, err := file.Stat()
					if err != nil {
						logrus.Errorf("Could not stat file %s. Err %v", path, err)
						continue
					}

					size := fileinfo.Size() >> 20
					//close file
					_ = file.Close()

					if size >= m.rotateAfterSize {
						logrus.Debugf("Rotate log at %s", t)
						if err := m.rotate(pid, path); err != nil {
							logrus.Errorf("Could not rotate logfile %s", path)
						}
					}
				}
			}
		}
	}()
}

func (m nginxLogRotate) rotate(pid int, path string) error {

	oldLog := fmt.Sprintf("%s.0", path)

	//mv access.log access.log.0
	if err := os.Rename(path, oldLog); err != nil {
		return errors.Wrap(err, "Could not rename log file")
	}

	//Signal nginx to reopen logs
	if err := syscall.Kill(pid, syscall.SIGUSR1); err != nil {
		return errors.Wrap(err, "Could not signal nginx")
	}

	//Write test
	if err := os.Remove(oldLog); err != nil {
		return errors.Wrap(err, "Could not delete old log file")
	}
	logrus.Infof("Removed %s", oldLog)

	return nil
}

func (m nginxExitHandler) HandleExit(exitCode int, pid int) int {

	if exitCode == int(syscall.SIGINT)+128 {
		logrus.Info("Nginx terminated successfully from a SIGINT")
		return 0
	}
	if exitCode == int(syscall.SIGTERM)+128 {
		logrus.Info("Nginx terminated successfully from a SIGTERM")
		return 0
	}
	return exitCode
}
