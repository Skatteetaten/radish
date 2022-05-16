package java

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/skatteetaten/radish/pkg/executor"
	"github.com/skatteetaten/radish/pkg/util"
)

type javaExitHandler struct {
}

type generatedJavaExecutor struct {
	javaExitHandler
}

//NewJavaExecutor :
func NewJavaExecutor() executor.Executor {
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
	argumentModificators := resolveArgumentModificators(os.Getenv)
	args, err := buildArgline(desc, os.LookupEnv, argumentModificators, util.ReadCGroupLimits())
	if err != nil {
		return nil, err
	}
	cmd := exec.Command("java", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd, nil
}

func (m *generatedJavaExecutor) BuildClasspath(radishDescriptor string) (string, error) {
	dat, err := ioutil.ReadFile(radishDescriptor)
	if err != nil {
		return "", err
	}
	desc, err := unmarshallDescriptor(bytes.NewBuffer(dat))
	if err != nil {
		return "", err
	}
	jars, err := createClasspath(desc.Data.Basedir, desc.Data.PathsToClassLibraries)
	if err != nil {
		return "", err
	}
	return strings.Join(jars, ":"), nil
}

func resolveArgumentModificators(env func(string) string) []ArgumentModificator {
	// This is set by the base image (wingnut<X>):
	majorVersion := env("JAVA_VERSION_MAJOR")
	switch majorVersion {
	case "8":
		logrus.Debug("Starting Java 8 process")
		return Java8ArgumentsModificators
	case "11":
		logrus.Debug("Starting Java 11 process")
		return Java11ArgumentsModificators
	case "17":
		logrus.Debug("Starting Java 17 process")
		return Java17ArgumentsModificators
	default:
		panic(fmt.Sprintf("Unsupported JAVA_VERSION_MAJOR: %s", majorVersion))
	}
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
