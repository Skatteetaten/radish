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
var configFilePath string
var outputFilePath string
var splunkIndex string
var podNamespace string
var appName string
var hostName string

var openshiftConfigPath string
var nginxPath string

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

	rootCmd.AddCommand(radish.GenerateNginxConfiguration)
	radish.GenerateNginxConfiguration.Flags().StringVarP(&openshiftConfigPath, "openshiftConfigPath", "", "", "path to the openshift.json file")
	radish.GenerateNginxConfiguration.Flags().StringVarP(&nginxPath, "nginxPath", "", "", "The nginxPath is the location (including file name) where the file is saved.")

	rootCmd.AddCommand(radish.GenerateSplunkStanzas)
	radish.GenerateSplunkStanzas.Flags().StringVarP(&templateFilePath, "templateFilePath", "t", "", "path of template. Will use default if not provided")

	radish.GenerateSplunkStanzas.Flags().StringVarP(&splunkIndex, "splunkIndex", "s", "", "SplunkIndex value - template variable, will attempt to use environment variable SPLUNK_INDEX if not set. ")
	radish.GenerateSplunkStanzas.Flags().StringVarP(&podNamespace, "podNamespace", "p", "", "PodNamespace value - template variable, will attempt to use environment variable POD_NAMESPACE if not set.")
	radish.GenerateSplunkStanzas.Flags().StringVarP(&appName, "appName", "a", "", "AppName value - template variable, will attempt to use environment variable APP_NAME if not set.")
	radish.GenerateSplunkStanzas.Flags().StringVarP(&hostName, "hostName", "n", "", "HostName value - template variable, will attempt to use environment variable HOST_NAME if not set.")

	radish.GenerateSplunkStanzas.Flags().StringVarP(&outputFilePath, "outputFilePath", "o", "", "path of output file")
	radish.GenerateSplunkStanzas.MarkFlagRequired("outputFilePath")

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
