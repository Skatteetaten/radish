package radish

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/skatteetaten/radish/pkg/startscript"
)

//GenerateStartScript : Use to generate startScript. Input params: configFilePath, outputFilePath
var GenerateStartScript = &cobra.Command{
	Use:   "generateStartScript",
	Short: "Use to generate startScript. Input params: configFilePath, outputFilePath",
	Long: `Requires 2 parameters: 
	1: Path to a file containing configuration json with classpath, mainclass, java options and application arguments. Example config:
	{
		"Classpath" : ["/app/lib/metrics.jar", "/app/lib/rt.jar", "/app/lib/spring.jar"],
		"JvmOptions" : "-Dfoo=bar",
		"MainClass" : "foo.bar.Main",	
		"ApplicationArgs" : "--logging.config=logback.xml"
	}

	2: Where to put the generated startscript file.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		//TODO: better validation of input params
		if len(args) < 1 {
			logrus.Error("No json config file parameter provided. Exiting..")
			os.Exit(2)
		}

		var configFilePath = args[0]
		var outputFilePath = args[1]

		success, err := startscript.GenerateStartscript(configFilePath, outputFilePath)
		if err != nil {
			logrus.Fatal(err)
			os.Exit(1) //TODO what to exit with?
		}

		if success {
			logrus.Infof("Startscript generated")
		}
	},
}

//RunPlaceholder :
var RunPlaceholder = &cobra.Command{
	Use:   "bash",
	Short: "Run radish with a placeholder process (tail -f /dev/null) for use with radish as entrypoint (DEPRECATED?)",
	Long:  `TODO help text`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info("Inside bash placeholder")
		comd := exec.Command("/usr/bin/tail", "-f", "/dev/null")
		comd.Start()
		pid := comd.Process.Pid
		logrus.Infof("pid: %d", pid)
		var wstatus syscall.WaitStatus
		syscall.Wait4(int(pid), &wstatus, 0, nil)
		exitCode := wstatus.ExitStatus()
		logrus.Infof("Exit code bash %d", exitCode)
		os.Exit(int(exitCode))

	},
}
