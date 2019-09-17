package nodejs

import (
	"encoding/json"
	"io"
)

//Type :
type Type struct {
	Type    string `envvar:"TYPE" default:"NodeJS"`
	Version string `envvar:"NODEJS_VERSION" default:""`
}

//DescriptorData :
type DescriptorData struct {
	HasNodeJSApplication bool   `envvar:"NODEJS_APP" default:"false"`
	AppVersion           string `envvar:"APP_VERSION" default:""`
	WebappPath           string `envvar:"WEB_APP_PATH" default:""`
	Path                 string `envvar:"PATH" default:""`
	NodeJSOverrides      string `envvar:"NODEJS_OVERRIDES" default:"{}"`
	Static               string `envvar:"STATIC" default:""`
	ExtraHeaders         string `envvar:"EXTRA_HEADERS" default:"{}"`
	SPA                  bool   `envvar:"SPA" default:"false"`
	ConfigurableProxy    bool   `envvar:"CONFIGURABLE_PROXY" default:"false"`
}

//Descriptor :
type Descriptor struct {
	Type
	Data DescriptorData
}

//UnmarshallDescriptor :
func UnmarshallDescriptor(buffer io.Reader) (Descriptor, error) {
	var data Descriptor
	err := json.NewDecoder(buffer).Decode(&data)
	return data, err
}
