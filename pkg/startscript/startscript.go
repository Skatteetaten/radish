package startscript

import (
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/skatteetaten/radish/pkg/util"
)

//TODO - put template in a file to ease tinkering?
var startscriptTemplate = `
exec radish {{.JvmOptions}} -cp "{{range $i, $value := .Classpath}}{{if $i}}:{{end}}{{$value}}{{end}}" $JAVA_OPTS {{.MainClass}} {{.ApplicationArgs}}
`

//Data : Struct for the required elements in the configuration json
type Data struct {
	Classpath       []string
	JvmOptions      string
	MainClass       string
	ApplicationArgs string
}

//GenerateStartscript : Use to generate startScript. Input params: configFilePath, outputFilePath
func GenerateStartscript(configFilePath string, outputFilePath string) (bool, error) {

	configJSONFromFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return false, errors.Wrap(err, "Failed to read config file")
	}

	var data Data
	data, err = unMarshalJSON(configJSONFromFile)
	if err != nil {
		return false, errors.Wrap(err, "Failed to unmarshal json config data")
	}

	fileWriter := util.NewFileWriter(outputFilePath)

	if err := fileWriter(newStartScript(data), "generated-start"); err != nil {
		return false, errors.Wrap(err, "Failed to write script")
	}

	return true, nil
}

func unMarshalJSON(dataJSON []byte) (Data, error) {
	var data Data
	err := json.Unmarshal(dataJSON, &data)
	return data, err
}

func newStartScript(data Data) util.WriterFunc {
	return util.NewTemplateWriter(
		data,
		"generatedStartScript",
		startscriptTemplate)
}
