package radish

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/skatteetaten/radish/pkg/splunk"
	"github.com/skatteetaten/radish/pkg/startscript"
)

//GenerateStartScript : Use to generate startScript. Input params: configFilePath, outputFilePath
var GenerateStartScript = &cobra.Command{
	Use:   "generateStartScript",
	Short: "Use to generate startScript. Input options: configFilePath, outputFilePath",
	Long: `Requires 2 parameters: 
	1. configFilePath - Path to a file containing configuration json with classpath, mainclass, java options and application arguments. Example config:
	{
		"Classpath" : ["/app/lib/metrics.jar", "/app/lib/rt.jar", "/app/lib/spring.jar"],
		"JvmOptions" : "-Dfoo=bar",
		"MainClass" : "foo.bar.Main",	
		"ApplicationArgs" : "--logging.config=logback.xml"
	}

	2. outputFilePath - Where to put the generated startscript file.
	`,
	Run: func(cmd *cobra.Command, args []string) {

		var configFilePath = cmd.Flag("configFilePath").Value.String()
		var outputFilePath = cmd.Flag("outputFilePath").Value.String()

		success, err := startscript.GenerateStartscript(configFilePath, outputFilePath)
		if err != nil {
			logrus.Fatalf("Startscript generation failed: %d", err)
			os.Exit(1) //TODO what to exit with?
		}

		if success {
			logrus.Infof("Startscript generated")
		}
	},
}

//GenerateSplunkStanzas : Use to generate Splunk stanzas. If a stanza template file is present, use it, if not, use default stanzas.
var GenerateSplunkStanzas = &cobra.Command{
	Use:   "generateSplunkStanzas",
	Short: "Use to generate Splunk stanzas. If a stanza template file is provided, use it, if not, use default stanzas.",
	Long: `For generating Splunk stanzas. 

Takes 3 parameters:

1. templateFilePath - path of a file containing a template. If empty string, the default template will be used.
	Default template:

		# --- start/stanza STDOUT
		[monitor://./logs/*.log]
		disabled = false
		followTail = 0
		sourcetype = log4j
		index = {{.SplunkIndex}}
		_meta = environment::{{.PodNamespace}} application::{{.AppName}} nodetype::openshift
		host = {{.HostName}}
		# --- end/stanza

		# --- start/stanza ACCESS_LOG
		[monitor://./logs/*.access]
		disabled = false
		followTail = 0
		sourcetype = access_combined
		index = {{.SplunkIndex}}
		_meta = environment::{{.PodNamespace}} application::{{.AppName}} nodetype::openshift
		host = {{.HostName}}
		# --- end/stanza

		# --- start/stanza GC LOG
		[monitor://./logs/*.gc]
		disabled = false
		followTail = 0
		sourcetype = gc_log
		index = {{.SplunkIndex}}
		_meta = environment::{{.PodNamespace}} application::{{.AppName}} nodetype::openshift
		host = {{.HostName}}
		# --- end/stanza

2. configFilePath - path to a JSON file containing template variables.

	Expected configuration file with 4 parameters as follows:

		{
			"SplunkIndex"  : "splunkIndex",
			"PodNamespace" : "podNameSpace",
			"AppName"      : "appName",
			"HostName"     : "hostName"
		}	

3. outputFilePath - path/name of the output file.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		template := ""
		if cmd.Flag("templateFilePath") != nil {
			template = cmd.Flag("templateFilePath").Value.String()
		}
		config := cmd.Flag("configFilePath").Value.String()
		output := cmd.Flag("outputFilePath").Value.String()

		success, err := splunk.GenerateStanzas(template, config, output)

		if err != nil {
			logrus.Fatalf("Splunk stanza generation failed: %d", err)
			os.Exit(1)
		}

		if success {
			logrus.Infof("Splunk Stanzas generated")
		}
	},
}

//RunPlaceholder :(DEPRECATED) Initally a way to start a container that used radish as entrypoint.
var RunPlaceholder = &cobra.Command{
	Use:   "placeholder",
	Short: "(DEPRECATED) Run radish with a placeholder process (tail -f /dev/null) for use with radish as entrypoint",
	Long:  `Will be removed`,
	Run: func(cmd *cobra.Command, args []string) {
		logrus.Info("Inside placeholder")
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
