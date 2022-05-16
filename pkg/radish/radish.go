package radish

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"

	"github.com/skatteetaten/radish/pkg/executor/java"
	"github.com/skatteetaten/radish/pkg/executor/nginx"
	"github.com/skatteetaten/radish/pkg/reaper"
	"github.com/skatteetaten/radish/pkg/signaler"
)

// RunRadish :
func RunRadish(args []string) {
	e := java.NewJavaExecutor()
	radishDescriptor, err := locateRadishDescriptor(args)
	if err != nil {
		logrus.Fatalf("Unable to load descriptor %s", err)
	}
	cmd, err := e.BuildCmd(radishDescriptor)
	if err != nil {
		logrus.Fatalf("Unable to start app %s", err)
	}
	logrus.Infof("Starting java with %s", strings.Join(cmd.Args, " "))
	err = cmd.Start()
	if err != nil {
		logrus.Fatalf("Error starting: %s", err)
	}
	reaper.Start()
	signal.Ignore(syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL)
	pid := cmd.Process.Pid
	signaler.Start(cmd.Process, findGraceTime())
	var wstatus syscall.WaitStatus
	syscall.Wait4(pid, &wstatus, 0, nil)
	logrus.Infof("Exit code %d", wstatus.ExitStatus())
	exitCode := e.HandleExit(wstatus.ExitStatus(), pid)
	os.Exit(exitCode)
}

// RunNginx :
func RunNginx(nginxConfigPath string, rotateLogsAfterSize, checkRotateAfter int) {
	e := nginx.NewNginxExecutor(rotateLogsAfterSize, checkRotateAfter, []string{"/u01/logs/nginx.access", "/u01/logs/nginx.log"})

	cmd := e.PrepareForNginxRun(nginxConfigPath)
	err := cmd.Start()
	if err != nil {
		logrus.Fatalf("Unable to start nginx: %v", err)
	}
	reaper.Start()
	signal.Ignore(syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGUSR1)
	pid := cmd.Process.Pid

	logrus.Infof("Started nginx with pid=%d", cmd.Process.Pid)

	e.StartLogRotate(pid)
	signaler.Start(cmd.Process, findGraceTime())

	var wstatus syscall.WaitStatus

	syscall.Wait4(pid, &wstatus, 0, nil)

	logrus.Infof("Exit code %d", wstatus.ExitStatus())

	if wstatus.Exited() && wstatus.ExitStatus() == 0 {
		logrus.Info("Nginx exited sucessfully")
	} else {
		logrus.Infof("Nginx terminated with exit code %d ", wstatus.ExitStatus())
	}
	os.Exit(wstatus.ExitStatus())
}

func findGraceTime() time.Duration {
	signalForward := os.Getenv("RADISH_SIGNAL_FORWARD_DELAY")
	if signalForward == "" {
		return 0
	}
	sf, err := strconv.Atoi(signalForward)
	if err != nil {
		logrus.Warnf("Could not parse %s to an integer (%s). Signal forward delay is 0", signalForward, err)
		return 0
	}

	return time.Duration(int64(sf) * int64(time.Second))
}

// PrintRadishCP :
func PrintRadishCP(args []string) {
	e := java.NewJavaExecutor()
	radishDescriptor, err := locateRadishDescriptor(args)
	if err != nil {
		logrus.Fatalf("Unable to load descriptor %s", err)
	}
	cp, err := e.BuildClasspath(radishDescriptor)
	if err != nil {
		logrus.Fatalf("Failed to build classpath %s", err)
	}
	fmt.Print(cp)
}

func locateRadishDescriptor(args []string) (string, error) {
	if len(args) > 0 {
		_, err := os.Stat(args[0])
		if err == nil {
			return args[0], nil
		}
		return "", errors.Wrapf(err, "Error reading %s", args[0])

	}
	descriptor, exists := os.LookupEnv("RADISH_DESCRIPTOR")
	if exists {
		return descriptor, nil
	}
	if _, err := os.Stat("/u01/app/radish.json"); err == nil {
		return "/u01/app/radish.json", nil
	}
	if _, err := os.Stat("/radish.json"); err == nil {
		return "/radish.json", nil
	}
	return "", errors.New("No radish descriptor found")
}

// GenerateNginxConfiguration :
func GenerateNginxConfiguration(openshiftConfigPath string, nginxPath string) error {
	return nginx.GenerateNginxConfiguration(openshiftConfigPath, nginxPath)
}
