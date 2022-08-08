package cmd

import (
	"fmt"
	"os"

	"strings"

	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/radish/cmd/radish"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "radish",
	Short: "Radish CLI",
	Long:  `Radish CLI`,
}

var templateFilePath string
var outputFilePath string
var splunkIndex string
var podNamespace string
var appName string
var hostName string

var openshiftConfigPath string
var nginxPath string

var mainJavascriptFile string
var stdoutLogLocation string
var stdoutLogFile string

//Execute :
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		logrus.Info(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(radish.RunJava)
	rootCmd.AddCommand(radish.PrintClasspath)

	rootCmd.AddCommand(radish.GenerateNginxConfiguration)
	radish.GenerateNginxConfiguration.Flags().StringVarP(&openshiftConfigPath, "radishConfigPath", "", "", "path to the radish config file")
	radish.GenerateNginxConfiguration.Flags().StringVarP(&nginxPath, "nginxPath", "", "", "The nginxPath is the location (including file name) where the file is saved.")

	rootCmd.AddCommand(radish.RunNginx)
	radish.RunNginx.Flags().StringVarP(&nginxPath, "nginxPath", "", "", "The nginxPath is the location (including file name) where the config file is stored.")
	radish.RunNginx.Flags().Int("rotateLogsAfterSize", 50, "Rotate logs when log size is above this value. Value is in MB")
	radish.RunNginx.Flags().Int("checkRotateAfter", 1000, "The interval in which we check log rotation")

	rootCmd.AddCommand(radish.RunNodeJS)
	radish.RunNodeJS.Flags().StringVarP(&mainJavascriptFile, "mainJavascriptFile", "", "", "The file name of the nodeJS program to run")
	radish.RunNodeJS.Flags().StringVarP(&stdoutLogLocation, "stdoutLogLocation", "", "/u01/logs", "Where the log is put - default /u01/logs")
	radish.RunNodeJS.Flags().StringVarP(&stdoutLogFile, "stdoutLogFile", "", "nodejs_stdout.log", "The file name for the file the nodejs stdout log ends up in. Default nodejs_stdout.log")
	radish.RunNodeJS.Flags().Int("stdoutFileRotateSize", 50, "The maximum size of the log file before log rotation - default max file size is 50MB")

	rootCmd.AddCommand(radish.GenerateEnvScript)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	//RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.architect.yaml)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//RootCmd.Flags().BoolP("verbose", "v", false, "Verbose logging")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if strings.ToUpper(os.Getenv("DEBUG")) == "TRUE" || strings.ToUpper(os.Getenv("RADISH_DEBUG")) == "TRUE" {
		logrus.SetLevel(logrus.DebugLevel)
	}
}
