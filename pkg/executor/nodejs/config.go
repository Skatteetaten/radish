package nodejs

import (
	"encoding/json"
	"io"
)

//Type :
type Type struct {
	Type    string `json:"Type"`
	Version string `json:"Version"`
}

//DescriptorData :
type DescriptorData struct {
	HasNodeJSApplication bool              `json:"HasNodeJSApplication"`
	AppVersion           string            `json:"AppVersion"`
	WebappPath           string            `json:"WebappPath"`
	Path                 string            `json:"Path"`
	NodeJSOverrides      map[string]string `json:"NodeJSOverrides"`
	Static               string            `json:"Static"`
	ExtraHeaders         map[string]string `json:"ExtraHeaders"`
	SPA                  bool              `json:"SPA"`
	ConfigurableProxy    bool              `json:"ConfigurableProxy"`
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
