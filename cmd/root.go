package cmd

import (
	"fmt"
	"os"

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
	rootCmd.AddCommand(radish.RunPlaceholder)
	rootCmd.AddCommand(radish.GenerateStartScript)
	radish.GenerateStartScript.Flags().StringVarP(&ConfigFilePath, "configFilePath", "c", "", "path of json config file")
	radish.GenerateStartScript.MarkFlagRequired("configFilePath")
	radish.GenerateStartScript.Flags().StringVarP(&OutputFilePath, "outputFilePath", "o", "", "path of output file")
	radish.GenerateStartScript.MarkFlagRequired("outputFilePath")

	rootCmd.AddCommand(radish.GenerateSplunkStanzas)
	radish.GenerateSplunkStanzas.Flags().StringVarP(&TemplateFilePath, "templateFilePath", "t", "", "path of template. Will use default if not provided")
	radish.GenerateSplunkStanzas.Flags().StringVarP(&ConfigFilePath, "configFilePath", "c", "", "path of json config file")
	radish.GenerateSplunkStanzas.MarkFlagRequired("configFilePath")
	radish.GenerateSplunkStanzas.Flags().StringVarP(&OutputFilePath, "outputFilePath", "o", "", "path of output file")
	radish.GenerateSplunkStanzas.MarkFlagRequired("outputFilePath")

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
	//TODO
}
