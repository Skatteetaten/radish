package radish

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/radish/pkg/executor"
	"github.com/skatteetaten/radish/pkg/reaper"
	"github.com/skatteetaten/radish/pkg/signaler"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func RunRadish(args []string) {
	e := executor.NewJavaExecutor()
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
	signaler.Start(cmd.Process)
	var wstatus syscall.WaitStatus
	syscall.Wait4(int(pid), &wstatus, 0, nil)
	logrus.Infof("Exit code %d", wstatus.ExitStatus())
	exitCode := e.HandleExit(wstatus.ExitStatus(), pid)
	os.Exit(int(exitCode))
}

func locateRadishDescriptor(args []string) (string, error) {
	if len(args) > 0 {
		if _, err := os.Stat(args[0]); err == nil {
			return args[0], nil
		} else {
			return "", errors.Wrapf(err, "Error reading %s", args[0])
		}
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
