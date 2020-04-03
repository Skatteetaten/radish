package executor

import (
	"os/exec"
)

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

	/**
	Build classpath based on radish descriptor
	*/
	BuildClasspath(string) (string, error)
}

//TemplateInput :
type TemplateInput struct {
	Baseimage          string
	NginxOverrides     map[string]string
	Static             string
	SPA                bool
	ExtraStaticHeaders map[string]string
	Path               string
	HasProxyPass       bool
	ProxyPassHost      string
	ProxyPassPort      string
	Gzip               string
	Exclude            []string
	Locations          string
	WorkerConnections  string
	WorkerProcesses    string
	NotServingOnRoot   bool
}
