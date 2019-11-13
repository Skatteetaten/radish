package nginx

import (
	"encoding/json"
	"io"
)

//Docker :
type Docker struct {
	Maintainer string            `json:"maintainer"`
	Labels     map[string]string `json:"labels"`
}

//Web :
type Web struct {
	ConfigurableProxy bool     `json:"configurableProxy"`
	Nodejs            Nodejs   `json:"nodejs"`
	WebApp            WebApp   `json:"webapp"`
	Exclude           []string `json:"exclude"`
}

//Nodejs :
type Nodejs struct {
	Main      string            `json:"main"`
	Overrides map[string]string `json:"overrides"`
}

//WebApp :
type WebApp struct {
	Content         string            `json:"content"`
	Path            string            `json:"path"`
	DisableTryfiles bool              `json:"disableTryfiles"`
	Headers         map[string]string `json:"headers"`
}

//OpenshiftConfig :
type OpenshiftConfig struct {
	Docker Docker `json:"docker"`
	Web    Web    `json:"web"`
}

//UnmarshallOpenshiftConfig :
func UnmarshallOpenshiftConfig(buffer io.Reader) (OpenshiftConfig, error) {
	var data OpenshiftConfig
	err := json.NewDecoder(buffer).Decode(&data)
	return data, err
}
