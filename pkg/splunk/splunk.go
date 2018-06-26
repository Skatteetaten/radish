package splunk

//go:generate go-bindata -pkg=$GOPACKAGE -o=resources/bindata.go resources

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"

	splunkresources "github.com/skatteetaten/radish/pkg/splunk/resources"
	"github.com/skatteetaten/radish/pkg/util"
)

//Data : Struct for the required elements in the configuration json
type Data struct {
	SplunkIndex  string
	PodNamespace string
	AppName      string
	HostName     string
}

//GenerateStanzas :
func GenerateStanzas(templateFilePath string, configFilePath string, outputFilePath string) (bool, error) {

	stanzatemplate, err := readStanzasTemplate(templateFilePath)
	if err != nil {
		return false, errors.Wrapf(err, "Failed to read template file from ", templateFilePath)
	}

	configParams, err := readConfigFile(configFilePath)
	if err != nil {
		return false, errors.Wrapf(err, "Failed to read config from ", configFilePath)
	}

	fileWriter := util.NewFileWriter(outputFilePath)

	if err := fileWriter(newSplunkStanzas(string(stanzatemplate), configParams), outputFilePath); err != nil {
		return false, errors.Wrap(err, "Failed to write Splunk stanzas")
	}

	return true, nil
}

func readStanzasTemplate(templateFilePath string) ([]byte, error) {

	if len(templateFilePath) > 0 {
		stanzatemplate, err := ioutil.ReadFile(templateFilePath)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed to read template file from ", templateFilePath)
		}
		return stanzatemplate, nil
	}

	stanzatemplate, err := splunkresources.Asset("resources/default_stanzas_template")
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read default template file from bindata")
	}
	return stanzatemplate, nil

}

func readConfigFile(configFilePath string) (*Data, error) {
	configFromFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read config file from ", configFilePath)
	}

	data, err := unMarshalJSON(configFromFile)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to unmarshal Json from config file ")
	}

	return &data, nil
}

func unMarshalJSON(dataJSON []byte) (Data, error) {
	var data Data
	err := json.Unmarshal(dataJSON, &data)
	return data, err
}

func newSplunkStanzas(template string, data *Data) util.WriterFunc {
	return util.NewTemplateWriter(
		data,
		"generatedSplunkStanzas",
		template)
}
