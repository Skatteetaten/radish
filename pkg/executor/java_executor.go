package executor

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/radish/pkg/util"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
)

type javaExitHandler struct {
}

type generatedJavaExecutor struct {
	javaExitHandler
}

func NewJavaExecutor() Executor {
	return &generatedJavaExecutor{
		javaExitHandler: javaExitHandler{},
	}
}

func (m *generatedJavaExecutor) BuildCmd(radishDescriptor string) (*exec.Cmd, error) {
	dat, err := ioutil.ReadFile(radishDescriptor)
	if err != nil {
		return nil, err
	}
	desc, err := unmarshallDescriptor(bytes.NewBuffer(dat))
	if err != nil {
		return nil, err
	}
	args, err := buildArgline(desc, os.LookupEnv, util.ReadCGroupLimits())
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("java", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, nil
}

func (m *javaExitHandler) HandleExit(exitCode int, pid int) int {
	if exitCode == int(syscall.SIGABRT)+128 {
		logrus.Info("Java is out of memory! Bummer")
		printCoreFileToStdOut(pid)
		return exitCode
	}
	if exitCode == int(syscall.SIGINT)+128 {
		logrus.Info("Java terminated successfully from a SIGINT")
		return 0
	}
	if exitCode == int(syscall.SIGTERM)+128 {
		logrus.Info("Java terminated successfully from a SIGTERM")
		return 0
	}
	return exitCode
}

func printCoreFileToStdOut(pid int) {
	report := fmt.Sprintf("hs_err_pid%d.log", pid)
	crashReport, err := ioutil.ReadFile(report)
	if err != nil {
		logrus.Errorf("Error reading crash report %s", report)
		return
	}
	logrus.Info(string(crashReport))
}
