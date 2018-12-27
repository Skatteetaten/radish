package executor

import "os/exec"

//Executor :
type Executor interface {
	/**
	Build a command
	*/
	BuildCmd(string) (*exec.Cmd, error)

	/**
	Handles exit code and rewrites exit code
	*/
	HandleExit(exitCode int, pid int) int
}
