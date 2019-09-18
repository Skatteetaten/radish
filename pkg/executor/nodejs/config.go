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
	HasNodeJSApplication bool   `envvar:"RADISH_NODEJS_APP" default:"false"`
	AppVersion           string `envvar:"RADISH_APP_VERSION" default:""`
	WebappPath           string `envvar:"RADISH_WEB_APP_PATH" default:""`
	Path                 string `envvar:"RADISH_PATH" default:""`
	NodeJSOverrides      string `envvar:"RADISH_NODEJS_OVERRIDES" default:"{}"`
	Static               string `envvar:"RADISH_STATIC" default:""`
	ExtraHeaders         string `envvar:"RADISH_EXTRA_HEADERS" default:"{}"`
	SPA                  bool   `envvar:"RADISH_SPA" default:"false"`
	ConfigurableProxy    bool   `envvar:"RADISH_CONFIGURABLE_PROXY" default:"false"`
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
