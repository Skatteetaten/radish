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

var TemplateFilePath string
var ConfigFilePath string
var OutputFilePath string
var SplunkIndex string
var PodNamespace string
var AppName string
var HostName string

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

	rootCmd.AddCommand(radish.GenerateSplunkStanzas)
	radish.GenerateSplunkStanzas.Flags().StringVarP(&TemplateFilePath, "templateFilePath", "t", "", "path of template. Will use default if not provided")

	radish.GenerateSplunkStanzas.Flags().StringVarP(&SplunkIndex, "splunkIndex", "s", "", "SplunkIndex value - template variable, will attempt to use environment variable SPLUNK_INDEX if not set. ")
	radish.GenerateSplunkStanzas.Flags().StringVarP(&PodNamespace, "podNamespace", "p", "", "PodNamespace value - template variable, will attempt to use environment variable POD_NAMESPACE if not set.")
	radish.GenerateSplunkStanzas.Flags().StringVarP(&AppName, "appName", "a", "", "AppName value - template variable, will attempt to use environment variable APP_NAME if not set.")
	radish.GenerateSplunkStanzas.Flags().StringVarP(&HostName, "hostName", "n", "", "HostName value - template variable, will attempt to use environment variable HOST_NAME if not set.")

	radish.GenerateSplunkStanzas.Flags().StringVarP(&OutputFilePath, "outputFilePath", "o", "", "path of output file")
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
