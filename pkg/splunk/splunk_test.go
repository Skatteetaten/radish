package splunk

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

const customSplunkStanza string = `# --- start/stanza CUSTOMIZED
[monitor://./logs/customfolder/*.log]
disabled = false
followTail = 0
sourcetype = custom_source_type
index = {{.SplunkIndex}}
_meta = environment::{{.PodNamespace}} application::{{.AppName}} nodetype::openshift somemore::meta
host = {{.HostName}}
# --- end/stanza
`

const customSplunkStanzaNewFormatResult string = `# --- start/stanza test
[monitor://./logs/*.log]
disabled = false
followTail = 0
sourcetype = log4j
index = test

_meta = environment::podNamespace application::appName nodetype::openshift
host = hostName
# --- end/stanza

# --- start/stanza audit
[monitor://./logs/*audit]
disabled = false
followTail = 0
sourcetype = _json
index = audit

_meta = environment::podNamespace application::appName nodetype::openshift
host = hostName
# --- end/stanza

# --- start/stanza access
[monitor://./logs/*access]
disabled = false
followTail = 0
sourcetype = combined
index = access

_meta = environment::podNamespace application::appName nodetype::openshift
host = hostName
# --- end/stanza

`

func TestGenerateStanzasCustomFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "radishtest")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	stanzaFile := dir + "/mystanzefile.stanza"
	err = writeFile(customSplunkStanza, stanzaFile)
	assert.NoError(t, err)

	outputFileName := dir
	splunkIndex := "overrideSplunkIndex"
	podNamespace := "podNamespace"
	appName := "appName"
	hostName := "hostName"

	os.Setenv("SPLUNK_INDEX", "splunkIndex")
	os.Setenv("SPLUNK_BLACKLIST", "splunkBlacklist")
	os.Setenv("POD_NAMESPACE", podNamespace)
	os.Setenv("APP_NAME", appName)
	os.Setenv("HOSTNAME", hostName)

	err = GenerateStanzas(stanzaFile, splunkIndex, "", "", "", outputFileName)
	assert.NoError(t, err)
	stanzaFileOutput := readFile(outputFileName + "/application.splunk")
	assert.True(t, generalStanzaFormat(stanzaFileOutput, 1, t))
	assert.True(t, strings.Contains(stanzaFileOutput, "# --- start/stanza CUSTOMIZED"))
	assert.True(t, strings.Count(stanzaFileOutput, "index = "+splunkIndex) == 1)
}

func TestGenerateStanzasAll(t *testing.T) {
	dir, err := ioutil.TempDir("", "radishtest")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	outputFileName := dir
	splunkIndex := "splunkIndex"
	splunkAuditIndex := "audit-test"
	splunkAppDynamicsIndex := "monitor"
	splunkAtsIndex := "ats-index"
	splunkBlacklist := "splunkBlacklist"
	podNamespace := "podNamespace"
	appName := "appName"
	hostName := "hostName"

	// Standard test, most used
	os.Setenv("SPLUNK_INDEX", splunkIndex)
	os.Setenv("SPLUNK_AUDIT_INDEX", splunkAuditIndex)
	os.Setenv("SPLUNK_APPDYNAMICS_INDEX", splunkAppDynamicsIndex)
	os.Setenv("SPLUNK_ATS_INDEX", splunkAtsIndex)
	os.Setenv("SPLUNK_BLACKLIST", splunkBlacklist)
	os.Setenv("POD_NAMESPACE", podNamespace)
	os.Setenv("APP_NAME", appName)
	os.Setenv("HOSTNAME", hostName)

	err = GenerateStanzas("", "", "", "", "", outputFileName)
	assert.NoError(t, err)
	stanzaFileOutput := readFile(outputFileName + "/application.splunk")
	assert.True(t, generalStanzaFormat(stanzaFileOutput, 7, t))
	assert.True(t, strings.HasPrefix(stanzaFileOutput, "# --- start/stanza STDOUT"))
	assert.True(t, strings.Contains(stanzaFileOutput, "# --- start/stanza ACCESS_LOG"))
	assert.True(t, strings.Contains(stanzaFileOutput, "# --- start/stanza GC LOG"))
	assert.True(t, strings.Contains(stanzaFileOutput, "# --- start/stanza AUDIT"))
	assert.True(t, strings.Contains(stanzaFileOutput, "# --- start/stanza APPDYNAMICS"))
	assert.True(t, strings.Count(stanzaFileOutput, "# --- start/stanza ATS") == 2)
	assert.True(t, strings.Count(stanzaFileOutput, "index = "+splunkIndex) == 3)
	assert.True(t, strings.Count(stanzaFileOutput, "index = "+splunkAuditIndex) == 1)
	assert.True(t, strings.Count(stanzaFileOutput, "blacklist = "+splunkBlacklist) == 3)

	// But without AppDynamics
	os.Setenv("SPLUNK_AUDIT_INDEX", "")
	os.Setenv("SPLUNK_APPDYNAMICS_INDEX", "")
	os.Setenv("SPLUNK_ATS_INDEX", "")
	err = GenerateStanzas("", "", "", "", "", outputFileName)
	stanzaFileOutput = readFile(outputFileName + "/application.splunk")
	assert.True(t, generalStanzaFormat(stanzaFileOutput, 3, t))

	// Test "command line" options
	splunkIndex = "newIndex"
	podNamespace = "newNameSpace"
	appName = "newAppName"
	hostName = "newHostName"
	err = GenerateStanzas("", splunkIndex, podNamespace, appName, hostName, outputFileName)
	assert.NoError(t, err)
	stanzaFileOutput = readFile(outputFileName + "/application.splunk")
	// Just set for test function.
	os.Setenv("POD_NAMESPACE", podNamespace)
	os.Setenv("APP_NAME", appName)
	os.Setenv("HOSTNAME", hostName)
	assert.True(t, generalStanzaFormat(stanzaFileOutput, 3, t))
	assert.True(t, strings.Count(stanzaFileOutput, "index = "+splunkIndex) == 3)
}

func TestGenerateStanzasNoApp(t *testing.T) {
	dir, err := ioutil.TempDir("", "radishtest")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	outputFileName := dir
	splunkAuditIndex := "audit-test"
	splunkAppDynamicsIndex := "monitor-123"

	os.Setenv("SPLUNK_INDEX", "")
	os.Setenv("SPLUNK_AUDIT_INDEX", splunkAuditIndex)
	os.Setenv("SPLUNK_APPDYNAMICS_INDEX", splunkAppDynamicsIndex)
	os.Setenv("POD_NAMESPACE", "podNamespace")
	os.Setenv("APP_NAME", "appName")
	os.Setenv("HOSTNAME", "hostName")

	err = GenerateStanzas("", "", "", "", "", outputFileName)
	assert.NoError(t, err)
	stanzaFileOutput := readFile(outputFileName + "/application.splunk")
	assert.True(t, generalStanzaFormat(stanzaFileOutput, 2, t))
	assert.True(t, strings.HasPrefix(stanzaFileOutput, "# --- start/stanza AUDIT"))
	assert.True(t, strings.Contains(stanzaFileOutput, "# --- start/stanza APPDYNAMICS"))
	assert.True(t, strings.Contains(stanzaFileOutput, "index = "+splunkAppDynamicsIndex))
	assert.True(t, strings.Contains(stanzaFileOutput, "index = "+splunkAuditIndex))
}

func TestGenerateNoStanzas(t *testing.T) {
	dir, err := ioutil.TempDir("", "radishtest")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	outputFileName := dir
	os.Setenv("POD_NAMESPACE", "podNamespace")
	os.Setenv("APP_NAME", "appName")
	os.Setenv("HOSTNAME", "hostName")
	os.Setenv("SPLUNK_INDEX", "")
	os.Setenv("SPLUNK_AUDIT_INDEX", "")
	os.Setenv("SPLUNK_APPDYNAMICS_INDEX", "")

	err = GenerateStanzas("", "", "", "", "", outputFileName)
	assert.NoError(t, err)
	_, err = os.Stat(outputFileName + "/application.splunk")
	assert.True(t, os.IsNotExist(err))
}

func TestGenerateStanzasNewFormat(t *testing.T) {
	dir, err := ioutil.TempDir("", "radishtest")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	outputFileName := dir
	splunkAuditIndex := "audit-test"
	splunkAppDynamicsIndex := "monitor-123"

	os.Setenv("SPLUNK_INDEX", "")
	os.Setenv("SPLUNK_AUDIT_INDEX", splunkAuditIndex)
	os.Setenv("SPLUNK_APPDYNAMICS_INDEX", splunkAppDynamicsIndex)
	os.Setenv("POD_NAMESPACE", "podNamespace")
	os.Setenv("APP_NAME", "appName")
	os.Setenv("HOSTNAME", "hostName")
	os.Setenv("SPLUNK_APPLICATION_LOG_CONFIG", "[{\"index\": \"test\", \"pattern\": \"*.log\", \"sourcetype\": \"log4j\"},{\"index\": \"audit\", \"pattern\":\"*audit\", \"sourcetype\": \"_json\"},{\"index\": \"access\", \"pattern\": \"*access\", \"sourcetype\": \"combined\"}]")

	err = GenerateStanzas("", "", "", "", "", outputFileName)
	assert.NoError(t, err)
	stanzaFileOutput := readFile(outputFileName + "/application.splunk")
	assert.Equal(t, customSplunkStanzaNewFormatResult, stanzaFileOutput)
	assert.True(t, generalStanzaFormat(stanzaFileOutput, 3, t))
}

func generalStanzaFormat(stanzaFile string, entries int, t *testing.T) bool {
	hostName := os.Getenv("HOSTNAME")
	podNamespace := os.Getenv("POD_NAMESPACE")
	appName := os.Getenv("APP_NAME")

	returnValue := true
	returnValue = returnValue && countEntries("# --- start/stanza", stanzaFile, entries, t)
	returnValue = returnValue && countEntries("# --- end/stanza", stanzaFile, entries, t)
	returnValue = returnValue && countEntries("disabled = false", stanzaFile, entries, t)
	returnValue = returnValue && countEntries("followTail = 0", stanzaFile, entries, t)
	returnValue = returnValue && countEntries(fmt.Sprintf("_meta = environment::%s application::%s", podNamespace, appName), stanzaFile, entries, t)
	returnValue = returnValue && countEntries("[monitor://.", stanzaFile, entries, t)
	returnValue = returnValue && countEntries("host = "+hostName, stanzaFile, entries, t)
	return returnValue
}

func countEntries(pattern string, file string, expectedEntries int, t *testing.T) bool {
	c := strings.Count(file, pattern)
	assert.Equal(t, expectedEntries, c, "Wrong count of "+pattern+" in file "+file)
	return expectedEntries == c
}

func readFile(stanzaFile string) string {
	stanzaFileOutput, err := ioutil.ReadFile(stanzaFile)
	if err != nil {
		return ""
	}
	return string(stanzaFileOutput)
}

func writeFile(stanzaFile string, fileName string) error {
	b1 := []byte(stanzaFile)
	err := ioutil.WriteFile(fileName, b1, 0644)
	if err != nil {
		return err
	}
	return nil
}
