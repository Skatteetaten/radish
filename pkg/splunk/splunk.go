package splunk

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/plaid/go-envvar/envvar"
	"github.com/sirupsen/logrus"
	"github.com/skatteetaten/radish/pkg/util"
	"io/ioutil"
	"text/template"
)

const splunkAppLogConfigTemplate string = `# --- start/stanza {{.NameOfIndexInComment}}
[monitor://./logs/{{.Pattern}}]
disabled = false
followTail = 0
sourcetype = {{.SourceType}}
index = {{.IndexName}}
{{.SplunkBlacklist}}
_meta = environment::{{.PodNamespace}} application::{{.AppName}} nodetype::openshift
host = {{.HostName}}
# --- end/stanza
`

//Data : Struct for the required elements in the configuration json
type Data struct {
	SplunkIndex            string `envvar:"SPLUNK_INDEX" default:""`
	SplunkAtsIndex         string `envvar:"SPLUNK_ATS_INDEX" default:""`
	SplunkAuditIndex       string `envvar:"SPLUNK_AUDIT_INDEX" default:""`
	SplunkAppdynamicsIndex string `envvar:"SPLUNK_APPDYNAMICS_INDEX" default:""`
	SplunkBlacklist        string `envvar:"SPLUNK_BLACKLIST" default:""`
	PodNamespace           string `envvar:"POD_NAMESPACE" default:""`
	AppName                string `envvar:"APP_NAME" default:""`
	HostName               string `envvar:"HOSTNAME" default:""`
	SplunkAppLogConfig     string `envvar:"SPLUNK_APPLICATION_LOG_CONFIG" default:""`
}

type splunkAppLogConfigElement struct {
	Name            string
	Index           string
	Pattern         string
	SourceType      string
	SplunkBlacklist string
}

type templateInput struct {
	PodNamespace         string
	AppName              string
	HostName             string
	IndexName            string
	NameOfIndexInComment string
	Pattern              string
	SourceType           string
	SplunkBlacklist      string
}

type applicationIdentifier struct {
	AppName      string
	AppNamespace string
	Hostname     string
}

func newApplicationIdentifier(appNameFlag string, podNamespaceFlag string, hostNameFlag string, vars Data) (*applicationIdentifier, error) {

	var podNamespace string
	if podNamespaceFlag != "" {
		podNamespace = podNamespaceFlag
	} else if vars.PodNamespace != "" {
		podNamespace = vars.PodNamespace
	} else {
		return nil, errors.New("No PodNamespace present as flag or environment variable")
	}

	var appName string
	if appNameFlag != "" {
		appName = appNameFlag
	} else if vars.AppName != "" {
		appName = vars.AppName
	} else {
		return nil, errors.New("No AppName present as flag or environment variable")
	}

	var hostname string
	if hostNameFlag != "" {
		hostname = hostNameFlag
	} else if vars.HostName != "" {
		hostname = vars.HostName
	} else {
		return nil, errors.New("No HostName present as flag or environment variable")
	}
	return &applicationIdentifier{
		AppName:      appName,
		AppNamespace: podNamespace,
		Hostname:     hostname,
	}, nil

}

//GenerateStanzas :
func GenerateStanzas(templateFilePath string, splunkIndexFlag string,
	podNamespaceFlag string, appNameFlag string, hostNameFlag string, outputFilePath string) error {
	vars := Data{}
	if err := envvar.Parse(&vars); err != nil {
		logrus.Fatal(err)
	}

	applicationIdentifier, err := newApplicationIdentifier(appNameFlag, podNamespaceFlag, hostNameFlag, vars)
	if err != nil {
		return err
	}
	var splunkStanza string
	var splunkStanzaElems []splunkAppLogConfigElement

	if vars.SplunkAppLogConfig != "" { // Use external configuration
		err := json.Unmarshal([]byte(vars.SplunkAppLogConfig), &splunkStanzaElems)
		if err != nil {
			return errors.Wrap(err, "Unable to unmarshal the splunk configuration")
		}

	} else { // Default to internal configuration
		if splunkIndexFlag != "" {
			vars.SplunkIndex = splunkIndexFlag
		}

		if vars.SplunkBlacklist == "" {
			logrus.Debug("No SPLUNK_BLACKLIST env variable present")
		} else {
			vars.SplunkBlacklist = "blacklist = " + vars.SplunkBlacklist
		}

		if vars.SplunkIndex == "" {
			logrus.Debug("No SPLUNK_INDEX env variable present")
		} else {
			splunkStanzaElems = append(splunkStanzaElems, createApplicationSplunkStanza(applicationIdentifier, vars.SplunkIndex, vars.SplunkBlacklist)...)
		}

		if vars.SplunkAuditIndex == "" {
			logrus.Debug("No SPLUNK_AUDIT_INDEX env variable present")
		} else {
			splunkStanzaElems = append(splunkStanzaElems, createAuditSplunkStanza(applicationIdentifier, vars.SplunkAuditIndex)...)
		}

		if vars.SplunkAppdynamicsIndex == "" {
			logrus.Debug("No SPLUNK_APPDYNAMICS_INDEX env variable present")
		} else {
			splunkStanzaElems = append(splunkStanzaElems, createAppdynamicsStanza(applicationIdentifier, vars.SplunkAppdynamicsIndex)...)
		}

		if vars.SplunkAtsIndex == "" {
			logrus.Debug("No SPLUNK_ATS_INDEX env variable present")
		} else {
			splunkStanzaElems = append(splunkStanzaElems, createAtsSplunkStanza(applicationIdentifier, vars.SplunkAtsIndex)...)
		}

		if vars.SplunkBlacklist == "" {
			logrus.Debug("No SPLUNK_BLACKLIST env variable present")
		} else {
			vars.SplunkBlacklist = "blacklist = " + vars.SplunkBlacklist
		}

		if len(templateFilePath) > 0 {
			logrus.Infof("Using template %s to generate splunk stanza file.", templateFilePath)
			stanzatemplate, err := readStanzasTemplate(templateFilePath)
			if err != nil {
				return errors.Wrapf(err, "Failed to read template file from %s", templateFilePath)
			}
			splunkStanza = string(stanzatemplate)

			fileWriter := util.NewFileWriter(outputFilePath)

			if err := fileWriter(newSplunkStanzas(splunkStanza, vars), "application.splunk"); err != nil {
				return errors.Wrap(err, "Failed to write Splunk stanzas")
			}
			logrus.Infof("Wrote splunk stanza to %s", outputFilePath)
			return nil
		}
	}

	for _, element := range splunkStanzaElems {
		if element.Name == "" {
			element.Name = element.Index
		}

		d := templateInput{
			AppName:              applicationIdentifier.AppName,
			PodNamespace:         applicationIdentifier.AppNamespace,
			HostName:             applicationIdentifier.Hostname,
			IndexName:            element.Index,
			NameOfIndexInComment: element.Name,
			Pattern:              element.Pattern,
			SourceType:           element.SourceType,
			SplunkBlacklist:      element.SplunkBlacklist,
		}

		stanzaElement, err := newSplunkStanzaElement(splunkAppLogConfigTemplate, d)
		if stanzaElement == "" {
			return errors.Wrap(err, "Failed to write Splunk stanzas")
		}
		splunkStanza = splunkStanza + stanzaElement + "\n"
	}

	if splunkStanza == "" {
		logrus.Info("No Splunk stanza will be created.")
		return nil
	}

	fileWriter := util.NewFileWriter(outputFilePath)
	content := util.NewStringWriter(fileWriter, splunkStanza)
	if err := fileWriter(content, "application.splunk"); err != nil {
		return errors.Wrap(err, "Failed to write Splunk stanzas")
	}

	logrus.Infof("Wrote splunk stanza to %s", outputFilePath)
	return nil
}

func createAtsSplunkStanza(identifier *applicationIdentifier, index string) []splunkAppLogConfigElement {
	return []splunkAppLogConfigElement{
		newStanzaElement("ATS Custom", index, "ats/*.custom.xml", "ats:eval:xml", ""),
		newStanzaElement("ATS DEFAULT", index, "ats/*.default.xml", "evalevent_xml", ""),
	}
}

func createAppdynamicsStanza(identifier *applicationIdentifier, index string) []splunkAppLogConfigElement {
	return []splunkAppLogConfigElement{
		newStanzaElement("APPDYNAMICS", index, "appdynamics/.../*.log", "log4j", ""),
	}
}

func createAuditSplunkStanza(identifier *applicationIdentifier, index string) []splunkAppLogConfigElement {
	return []splunkAppLogConfigElement{
		newStanzaElement("AUDIT", index, "*.audit.json", "_json", ""),
	}
}

func createApplicationSplunkStanza(identifier *applicationIdentifier, index string, splunkBlacklist string) []splunkAppLogConfigElement {
	return []splunkAppLogConfigElement{
		newStanzaElement("STDOUT", index, "*.log", "log4j", splunkBlacklist),
		newStanzaElement("ACCESS_LOG", index, "*.access", "access_combined", splunkBlacklist),
		newStanzaElement("GC LOG", index, "*.gc", "gc_log", splunkBlacklist),
	}
}

func newStanzaElement(name, index, pattern, sourcetype, splunkBlacklist string) splunkAppLogConfigElement {
	return splunkAppLogConfigElement{
		Name:            name,
		Index:           index,
		Pattern:         pattern,
		SourceType:      sourcetype,
		SplunkBlacklist: splunkBlacklist,
	}
}

func readStanzasTemplate(templateFilePath string) ([]byte, error) {
	stanzatemplate, err := ioutil.ReadFile(templateFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read template file from %s", templateFilePath)
	}
	return stanzatemplate, nil
}

func newSplunkStanzas(template string, data Data) util.WriterFunc {
	return util.NewTemplateWriter(
		data,
		"generatedSplunkStanzas",
		template)
}

func newSplunkStanzaElement(stanzaTemplate string, data templateInput) (string, error) {
	t, err := template.New("parseTemplate").Parse(stanzaTemplate)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, data); err != nil {
		return "", err
	}

	return tpl.String(), nil
}
