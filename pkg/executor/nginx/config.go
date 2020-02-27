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
	ConfigurableProxy bool           `json:"configurableProxy"`
	Nodejs            Nodejs         `json:"nodejs"`
	WebApp            WebApp         `json:"webapp"`
	Gzip              nginxGzip      `json:"gzip"`
	Exclude           []string       `json:"exclude"`
	Locations         nginxLocations `json:"locations"`
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

type headers map[string]string

type nginxLocations map[string]*nginxLocation

type nginxLocation struct {
	Headers headers   `json:"headers"`
	Gzip    nginxGzip `json:"gzip"`
}

type nginxGzip struct {
	Use         string `json:"use"`
	UseStatic   string `json:"use_static"`
	MinLength   int    `json:"min_length"`
	Vary        string `json:"vary"`
	Proxied     string `json:"proxied"`
	Types       string `json:"types"`
	Disable     string `json:"disable"`
	HTTPVersion string `json:"http_version"`
}

//UnmarshallOpenshiftConfig :
func UnmarshallOpenshiftConfig(buffer io.Reader) (OpenshiftConfig, error) {
	var data OpenshiftConfig
	err := json.NewDecoder(buffer).Decode(&data)
	return data, err
}
