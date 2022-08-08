package radish

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"fmt"

	"github.com/skatteetaten/radish/pkg/auroraenv"
	"github.com/skatteetaten/radish/pkg/radish"
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

//GenerateNginxConfiguration : Use to generate Nginx configuration files.
var GenerateNginxConfiguration = &cobra.Command{
	Use:   "generateNginxConfiguration",
	Short: "Use to generate Nginx configuration files based on a Radish descriptor",
	Long: `For generating Nginx configuration files. 

Takes a number of flags:

1. radishConfigPath - Path to the openshift.json file in the container. The file content is used to extract data for the nginx.conf file.
	radish.json file example:

	{
		"docker": {
		  "maintainer": "Aurora OpenShift Utvikling <utvpaas@skatteetaten.no>",
		  "labels": {
			"io.k8s.description": "Demo application with React on Openshift.",
			"io.openshift.tags": "openshift,react,nodejs"
		  }
		},
		"web": {
			"configurableProxy": false,
			"nodejs": {
				"main": "api/server.js",
				"overrides": {
					"client_max_body_size": "10m"
				}
			},
			"webapp": {
			   "content": "build",
			   "path": "/web",
			   "disableTryfiles": false,
			   "headers": {
				  "SomeHeader": "SomeValue"
				}
			}
		}
	  }

2. nginxPath - This command will generate an nginx configuration file. The nginxPath is the location (including file name) where the file is saved. 

`,
	Run: func(cmd *cobra.Command, args []string) {
		openshiftConfigPath := ""
		if cmd.Flag("radishConfigPath") != nil {
			openshiftConfigPath = cmd.Flag("radishConfigPath").Value.String()
		}

		nginxPath := ""
		if cmd.Flag("nginxPath") != nil {
			nginxPath = cmd.Flag("nginxPath").Value.String()
		}

		err := radish.GenerateNginxConfiguration(openshiftConfigPath, nginxPath)
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

//RunNginx :
var RunNginx = &cobra.Command{
	Use:   "runNginx",
	Short: "Runs a Nginx process with radish",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		nginxPath := ""
		if cmd.Flag("nginxPath") != nil {
			nginxPath = cmd.Flag("nginxPath").Value.String()
		} else {
			logrus.Fatal("nginx config path not present")
		}

		rotateLogsAfterSize, err := cmd.Flags().GetInt("rotateLogsAfterSize")
		if err != nil {
			logrus.Fatalf("Could not read value rotateLogsAfterSize: %v", err)
		}

		checkRotateAfter, err := cmd.Flags().GetInt("checkRotateAfter")
		if err != nil {
			logrus.Fatalf("Could not read value checkRotateAfter: %v", err)
		}

		radish.RunNginx(nginxPath, rotateLogsAfterSize, checkRotateAfter)
	},
}

//RunNodeJS :
var RunNodeJS = &cobra.Command{
	Use:   "runNodeJS",
	Short: "Runs a NodeJS process with radish",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		mainJavascriptFile := ""
		if cmd.Flag("mainJavascriptFile") != nil {
			mainJavascriptFile = cmd.Flag("mainJavascriptFile").Value.String()
		} else {
			logrus.Fatal("mainJavascriptFile not present")
		}

		stdoutLogLocation := ""
		if cmd.Flag("stdoutLogLocation") != nil {
			stdoutLogLocation = cmd.Flag("stdoutLogLocation").Value.String()
		} else {
			stdoutLogLocation = "/u01/logs"
		}

		stdoutLogFile := ""
		if cmd.Flag("stdoutLogFile") != nil {
			stdoutLogFile = cmd.Flag("stdoutLogFile").Value.String()
		} else {
			logrus.Fatal("stdoutLogFile not present")
		}

		stdoutFileRotateSize := 50
		if cmd.Flag("stdoutFileRotateSize") != nil {
			stdoutFileRotateSizeStr := cmd.Flag("stdoutFileRotateSize").Value.String()
			stdoutFileRotateSizeInt, err := strconv.Atoi(stdoutFileRotateSizeStr)
			if err != nil {
				logrus.Fatal("stdoutFileRotateSize is not an integer")
			}
			stdoutFileRotateSize = stdoutFileRotateSizeInt
		}
		radish.RunNodeJS(mainJavascriptFile, stdoutLogLocation, stdoutLogFile, stdoutFileRotateSize)
	},
}

//PrintClasspath :
var PrintClasspath = &cobra.Command{
	Use:   "printCP",
	Short: "Prints complete classpath Radish will use with java application",
	Long:  `Prints complete classpath Radish will use with java application`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		radish.PrintRadishCP(args)
	},
}
