package radish

import (
	"os"
	"os/exec"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/skatteetaten/radish/pkg/auroraenv"
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

//SetAuroraEnv : Use to set environment variables from appropriate properties files, based on app- and aurora versions.
var SetAuroraEnv = &cobra.Command{
	Use:   "setAuroraEnv",
	Short: "Use to set environment variables from appropriate properties files, based on app- and aurora versions.",
	Long: `For setting environment variables based on properties files. 
	Which properties files is deduced from environment variables APP_VERSION and AURORA_VERSION.
	The environment variable HOME is also required, as the base folder for all operations.
	This command is looking for .properties files in $HOME/config/{secrets, configmaps}
	`,
	Run: func(cmd *cobra.Command, args []string) {
		success, err := auroraenv.SetAuroraEnv()
		if err != nil {
			logrus.Fatalf("Setting Aurora environment variables failed: %s", err)
			os.Exit(1) //TODO what to exit with?
		}
		if success {
			logrus.Infof("Aurora environment variables set")
		}

	},
}

//GenerateSplunkStanzas : Use to generate Splunk stanzas. If a stanza template file is present, use it, if not, use default stanzas.
var GenerateSplunkStanzas = &cobra.Command{
	Use:   "generateSplunkStanzas",
	Short: "Use to generate Splunk stanzas. If a stanza template file is provided, use it, if not, use default stanzas.",
	Long: `For generating Splunk stanzas. 

Takes a number of flags:

1. templateFilePath - optional - path of a file containing a template. If not provided, the default template will be used.
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

2. splunkIndex - optional - template variable. Overrides environment variable SPLUNK_INDEX

3. podNamespace - optional - template variable. Overrides environment variable POD_NAMESPACE

4. appName - optional - template variable. Overrides environment variable APP_NAME

5. hostName - optional - template variable. Overrides environment variable HOST_NAME

6. outputFilePath - path/name of the output file.

For the template variables, they are only semi-optional. Radish generateSplunkStanzas will fail 
if no environment variable or the corresponding flag is set for any of the four template variables. 
In other words, if the flag is not set, then the environment variable must exist.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		template := ""
		if cmd.Flag("templateFilePath") != nil {
			template = cmd.Flag("templateFilePath").Value.String()
		}

		output := cmd.Flag("outputFilePath").Value.String()
		splunkIndex := cmd.Flag("splunkIndex").Value.String()
		podNamespace := cmd.Flag("podNamespace").Value.String()
		appName := cmd.Flag("appName").Value.String()
		hostName := cmd.Flag("hostName").Value.String()

		success, err := splunk.GenerateStanzas(template, splunkIndex, podNamespace, appName, hostName, output)

		if err != nil {
			logrus.Fatalf("Splunk stanza generation failed: %s", err)
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
