package nginx

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

//Executor :
type Executor interface {
	PrepareForNginxRun(nginxConfigPath string) *exec.Cmd
	StartLogRotate(pid int)
}

type nginxExecutor struct {
	nginxExitHandler
	nginxLogRotate
}

type nginxExitHandler struct {
}

type nginxLogRotate struct {
	paths            []string
	rotateAfterSize  int
	checkRotateAfter int
}

//NewNginxExecutor :
func NewNginxExecutor(rotateAfterSize int, checkRotateAfter int, logfiles []string) Executor {
	return nginxExecutor{
		nginxExitHandler{},
		nginxLogRotate{
			paths:            logfiles,
			rotateAfterSize:  rotateAfterSize,
			checkRotateAfter: checkRotateAfter,
		},
	}
}

func (m nginxExecutor) PrepareForNginxRun(nginxConfigPath string) *exec.Cmd {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("nginx -g 'daemon off;' -c %s", nginxConfigPath))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func (m nginxLogRotate) StartLogRotate(pid int) {
	ticker := time.NewTicker(time.Duration(m.checkRotateAfter) * time.Millisecond)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				//Check file size of all files
				for _, path := range m.paths {

					fileinfo, err := os.Stat(path)
					if err != nil {
						if os.IsNotExist(err) {
							continue
						}
						logrus.Errorf("Could not stat file %s. Err %v", path, err)
						continue
					}

					sizeInMb := fileinfo.Size() >> 20
					//close file

					if int(sizeInMb) >= m.rotateAfterSize {
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

	var extension = filepath.Ext(path)
	var base = path[0 : len(path)-len(extension)]
	oldLog := fmt.Sprintf("%s.1%s", base, extension)

	//mv access.log access.1.log
	if err := os.Rename(path, oldLog); err != nil {
		return errors.Wrap(err, "Could not rename log file")
	}

	//Signal nginx to reopen logs
	if err := syscall.Kill(pid, syscall.SIGUSR1); err != nil {
		return errors.Wrap(err, "Could not signal nginx")
	}

	return nil
}
