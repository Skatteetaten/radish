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
	AppVersion        string            `json:"AppVersion"`
	Labels            map[string]string `json:"Labels"`
	Maintainer        string            `json:"Maintainer"`
	WebappPath        string            `json:"WebappPath"`
	Path              string            `json:"Path"`
	NodeJSMain        string            `json:"NodeJSMain"`
	NodeJSOverrides   map[string]string `json:"NodeJSOverrides"`
	Static            string            `json:"Static"`
	ExtraHeaders      map[string]string `json:"ExtraHeaders"`
	SPA               bool              `json:"SPA"`
	ConfigurableProxy bool              `json:"ConfigurableProxy"`
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
