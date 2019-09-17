package radish

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"fmt"

	"github.com/skatteetaten/radish/pkg/auroraenv"
	"github.com/skatteetaten/radish/pkg/radish"
	"github.com/skatteetaten/radish/pkg/splunk"
)

//GenerateEnvScript : Use to set environment variables from appropriate properties files, based on app- and aurora versions.
var GenerateEnvScript = &cobra.Command{
	Use:   "generateEnvScript",
	Short: "Use to set environment variables from appropriate properties files, based on app- and aurora versions.",
	Long: `For setting environment variables based on properties files. 
	Running this command will print a number of export statements that can be eval'ed.

	Example usage: eval $(radish generateEnvScript)

	Which properties files is deduced from environment variables APP_VERSION and AURORA_VERSION.
	The environment variable HOME is also required, as the base folder for all operations.
	This command is looking for .properties files in $HOME/config/{secrets, configmaps}
	`,
	Run: func(cmd *cobra.Command, args []string) {
		//We set output to stderr since we eval from stdout
		logrus.SetOutput(os.Stderr)
		shellscript, err := auroraenv.GenerateEnvScript()
		if err != nil {
			logrus.Fatalf("Setting Aurora environment variables failed: %s", err)
		} else {
			fmt.Println(shellscript)
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

		err := splunk.GenerateStanzas(template, splunkIndex, podNamespace, appName, hostName, output)

		if err != nil {
			logrus.Fatalf("Splunk stanza generation failed: %s", err)
			os.Exit(1)
		}
	},
}

//GenerateNginxConfiguration : Use to generate Nginx configuration files.
var GenerateNginxConfiguration = &cobra.Command{
	Use:   "generateNginxConfiguration",
	Short: "Use to generate Nginx configuration files based on a Radish descriptor",
	Long: `For generating Nginx configuration files. 

Takes a number of flags:

You must include either the radishDescriptorPath or the radishDescriptor.

1. radishDescriptorPath - optional - Path to the descriptor file. The file is a JSON document which holds the configuration values for the nginx.conf file.
	Descriptor file example:

	{
	  "Type": "NodeJSDescriptor",
	  "Version": "1",
	  "Data": {
	    "AppVersion": "1.0",
	    "WebappPath": "/web",
	    "Path": "/",    
	    "NodeJSOverrides": {"client_max_body_size":"10m"},
	    "Static": "Dummy",      
	    "SPA": false,
	    "ConfigurableProxy": true
	  }
	}

2. radishDescriptor - optional - The descriptor file content

2. nginxPath - This command will generate an nginx configuration file. The nginxPath is the location (including file name) where the file is saved. 

`,
	Run: func(cmd *cobra.Command, args []string) {
		radishDescriptorPath := ""
		if cmd.Flag("radishDescriptorPath") != nil {
			radishDescriptorPath = cmd.Flag("radishDescriptorPath").Value.String()
		}

		radishDescriptor := ""
		if cmd.Flag("radishDescriptor") != nil {
			radishDescriptor = cmd.Flag("radishDescriptor").Value.String()
		}

		nginxPath := ""
		if cmd.Flag("nginxPath") != nil {
			nginxPath = cmd.Flag("nginxPath").Value.String()
		}

		err := radish.GenerateNginxConfiguration(radishDescriptorPath, radishDescriptor, nginxPath)
		if err != nil {
			logrus.Fatalf("Nginx config generation failed: %s", err)
			os.Exit(1)
		}
	},
}

//RunJava :
var RunJava = &cobra.Command{
	Use:   "runJava",
	Short: "Runs a Java process with Radish",
	Long:  `Runs a Java process with Radish. It automatically detects CGroup limits and some common flags`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		radish.RunRadish(args)
	},
}
