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
	_ "text/template"
)

const applicationSplunkStanza string = `# --- start/stanza STDOUT
[monitor://./logs/*.log]
disabled = false
followTail = 0
sourcetype = log4j
index = {{.SplunkIndex}}
{{.SplunkBlacklist}}
_meta = environment::{{.PodNamespace}} application::{{.AppName}} nodetype::openshift
host = {{.HostName}}
# --- end/stanza

# --- start/stanza ACCESS_LOG
[monitor://./logs/*.access]
disabled = false
followTail = 0
sourcetype = access_combined
index = {{.SplunkIndex}}
{{.SplunkBlacklist}}
_meta = environment::{{.PodNamespace}} application::{{.AppName}} nodetype::openshift
host = {{.HostName}}
# --- end/stanza

# --- start/stanza GC LOG
[monitor://./logs/*.gc]
disabled = false
followTail = 0
sourcetype = gc_log
index = {{.SplunkIndex}}
{{.SplunkBlacklist}}
_meta = environment::{{.PodNamespace}} application::{{.AppName}} nodetype::openshift
host = {{.HostName}}
# --- end/stanza
`

const atsSplunkStanza string = `# --- start/stanza ATS CUSTOM
[monitor://./logs/ats/*.custom.xml]
disabled = false
followTail = 0
sourcetype = ats:eval:xml
index = {{.SplunkAtsIndex}}
_meta = environment::{{.PodNamespace}} application::{{.AppName}} nodetype::openshift
host = {{.HostName}}
# --- end/stanza

# --- start/stanza ATS DEFAULT
[monitor://./logs/ats/*.default.xml]
disabled = false
followTail = 0
sourcetype = evalevent_xml
index = {{.SplunkAtsIndex}}
_meta = environment::{{.PodNamespace}} application::{{.AppName}} nodetype::openshift
host = {{.HostName}}
# --- end/stanza
`

const auditSplunkStanza string = `# --- start/stanza AUDIT
[monitor://./logs/*.audit.json]
disabled = false
followTail = 0
sourcetype = _json
index = {{.SplunkAuditIndex}}
_meta = environment::{{.PodNamespace}} application::{{.AppName}} nodetype::openshift logtype::audit
host = {{.HostName}}
# --- end/stanza
`

const appdynamicsSplunkStanza string = `# --- start/stanza APPDYNAMICS
[monitor://./logs/appdynamics/*.log]
disabled = false
followTail = 0
sourcetype = log4j
index = {{.SplunkAppdynamicsIndex}}
_meta = environment::{{.PodNamespace}} application::{{.AppName}} nodetype::openshift
host = {{.HostName}}
# --- end/stanza
`

const splunkAppLogConfigTemplate string = `# --- start/stanza AUDIT
[monitor://{{.Pattern}}]
disabled = false
followTail = 0
sourcetype = {{.SourceType}}
index = {{.Index}}
_meta = environment::{{.PodNamespace}} application::{{.AppName}} nodetype::openshift logtype::audit
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

type SplunkAppLogConfigElement struct {
	Index        string
	Pattern      string
	SourceType   string
	PodNamespace string
	AppName      string
	HostName     string
}

//GenerateStanzas :
func GenerateStanzas(templateFilePath string, splunkIndexFlag string,
	podNamespaceFlag string, appNameFlag string, hostNameFlag string, outputFilePath string) error {
	vars := Data{}
	if err := envvar.Parse(&vars); err != nil {
		logrus.Fatal(err)
	}

	var splunkStanza string

	if vars.SplunkAppLogConfig != "" { // Use external configuration
		var splunkAppLogConfigList []SplunkAppLogConfigElement
		json.Unmarshal([]byte(vars.SplunkAppLogConfig), &splunkAppLogConfigList)

		for _, element := range splunkAppLogConfigList {
			stanzaElement, err := newSplunkStanzaElement(splunkAppLogConfigTemplate, element)
			if stanzaElement == "" {
				return errors.Wrap(err, "Failed to write Splunk stanzas")
			}

			splunkStanza = splunkStanza + stanzaElement + "\n"
		}

		fileWriter := util.NewFileWriter(outputFilePath)
		content := util.NewStringWriter(fileWriter, splunkStanza)
		if err := fileWriter(content, "application.splunk"); err != nil {
			return errors.Wrap(err, "Failed to write Splunk stanzas")
		}

	} else { // Default to internal configuration
		if splunkIndexFlag != "" {
			vars.SplunkIndex = splunkIndexFlag
		}

		if vars.SplunkIndex == "" {
			logrus.Debug("No SPLUNK_INDEX env variable present")
		} else {
			splunkStanza = applicationSplunkStanza
		}

		if vars.SplunkAuditIndex == "" {
			logrus.Debug("No SPLUNK_AUDIT_INDEX env variable present")
		} else {
			if splunkStanza != "" {
				splunkStanza = splunkStanza + "\n"
			}
			splunkStanza = splunkStanza + auditSplunkStanza
		}

		if vars.SplunkAppdynamicsIndex == "" {
			logrus.Debug("No SPLUNK_APPDYNAMICS_INDEX env variable present")
		} else {
			if splunkStanza != "" {
				splunkStanza = splunkStanza + "\n"
			}
			splunkStanza = splunkStanza + appdynamicsSplunkStanza
		}

		if vars.SplunkAtsIndex == "" {
			logrus.Debug("No SPLUNK_ATS_INDEX env variable present")
		} else {
			if splunkStanza != "" {
				splunkStanza = splunkStanza + "\n"
			}
			splunkStanza = splunkStanza + atsSplunkStanza
		}

		if vars.SplunkBlacklist == "" {
			logrus.Debug("No SPLUNK_BLACKLIST env variable present")
		} else {
			vars.SplunkBlacklist = "blacklist = " + vars.SplunkBlacklist
		}

		if splunkStanza == "" {
			logrus.Info("No Splunk stanza will be created.")
			return nil
		}

		if podNamespaceFlag != "" {
			vars.PodNamespace = podNamespaceFlag
		} else if vars.PodNamespace == "" {
			return errors.New("No PodNamespace present as flag or environment variable")
		}

		if appNameFlag != "" {
			vars.AppName = appNameFlag
		} else if vars.AppName == "" {
			return errors.New("No AppName present as flag or environment variable")
		}

		if hostNameFlag != "" {
			vars.HostName = hostNameFlag
		} else if vars.HostName == "" {
			return errors.New("No HostName present as flag or environment variable")
		}

		if len(templateFilePath) > 0 {
			logrus.Infof("Using template %s to generate splunk stanza file.", templateFilePath)
			stanzatemplate, err := readStanzasTemplate(templateFilePath)
			if err != nil {
				return errors.Wrapf(err, "Failed to read template file from %s", templateFilePath)
			}
			splunkStanza = string(stanzatemplate)
		}

		fileWriter := util.NewFileWriter(outputFilePath)
		if err := fileWriter(newSplunkStanzas(splunkStanza, vars), "application.splunk"); err != nil {
			return errors.Wrap(err, "Failed to write Splunk stanzas")
		}
	}

	logrus.Infof("Wrote splunk stanza to %s", outputFilePath)
	return nil
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

func newSplunkStanzaElement(stanzaTemplate string, data SplunkAppLogConfigElement) (string, error) {
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
