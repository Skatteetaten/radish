package main

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/radish/cmd"
	"github.com/skatteetaten/radish/pkg/radish"
)

func main() {

	logrus.Infof("Started Radish with args: ", os.Args)

	// The arguments contain java option "-cp", we assume we should run the java executor
	if strings.Contains(strings.Join(os.Args, " "), "-cp") {
		radish.RunRadish(os.Args)
	} else {
		cmd.Execute()
	}
}
