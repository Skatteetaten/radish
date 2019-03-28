package splunk

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"strings"
	"testing"
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
	os.Setenv("POD_NAMESPACE", podNamespace)
	os.Setenv("APP_NAME", appName)
	os.Setenv("HOSTNAME", hostName)

	err = GenerateStanzas(stanzaFile, splunkIndex, "", "", "", outputFileName)
	assert.NoError(t, err)
	stanzaFileOutput := readFile(outputFileName + "/application.splunk")
	assert.True(t, generalStanzaFormat(stanzaFileOutput, 1))
	assert.True(t, strings.Contains(stanzaFileOutput, "# --- start/stanza CUSTOMIZED"))
	assert.True(t, strings.Count(stanzaFileOutput, "index = "+splunkIndex) == 1)
	t.Log(stanzaFileOutput)
}

func TestGenerateStanzasAll(t *testing.T) {
	dir, err := ioutil.TempDir("", "radishtest")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	outputFileName := dir
	splunkIndex := "splunkIndex"
	splunkAuditIndex := "audit-test"
	splunkAppDynamicsIndex := "monitor"
	podNamespace := "podNamespace"
	appName := "appName"
	hostName := "hostName"

	// Standard test, most used
	os.Setenv("SPLUNK_INDEX", splunkIndex)
	os.Setenv("SPLUNK_AUDIT_INDEX", splunkAuditIndex)
	os.Setenv("SPLUNK_APPDYNAMICS_INDEX", splunkAppDynamicsIndex) 
	os.Setenv("POD_NAMESPACE", podNamespace)
	os.Setenv("APP_NAME", appName)
	os.Setenv("HOSTNAME", hostName)

	t.Log(outputFileName)
	err = GenerateStanzas("", "", "", "", "", outputFileName)
	assert.NoError(t, err)
	stanzaFileOutput := readFile(outputFileName + "/application.splunk")
	assert.True(t, generalStanzaFormat(stanzaFileOutput, 5))
	assert.True(t, strings.HasPrefix(stanzaFileOutput, "# --- start/stanza STDOUT"))
	assert.True(t, strings.Contains(stanzaFileOutput, "# --- start/stanza ACCESS_LOG"))
	assert.True(t, strings.Contains(stanzaFileOutput, "# --- start/stanza GC LOG"))
	assert.True(t, strings.Contains(stanzaFileOutput, "# --- start/stanza AUDIT"))
	assert.True(t, strings.Contains(stanzaFileOutput, "# --- start/stanza APPDYNAMICS"))
	assert.True(t, strings.Count(stanzaFileOutput, "index = "+splunkIndex) == 3)
	assert.True(t, strings.Count(stanzaFileOutput, "index = "+splunkAuditIndex) == 1)
	t.Log(stanzaFileOutput)

	// But without AppDynamics
	os.Setenv("SPLUNK_AUDIT_INDEX", "")
	os.Setenv("SPLUNK_APPDYNAMICS_INDEX", "") 
	err = GenerateStanzas("", "", "", "", "", outputFileName)
	stanzaFileOutput = readFile(outputFileName + "/application.splunk")
	assert.True(t, generalStanzaFormat(stanzaFileOutput, 3))

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
	assert.True(t, generalStanzaFormat(stanzaFileOutput, 3))
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
	assert.True(t, generalStanzaFormat(stanzaFileOutput, 2))
	assert.True(t, strings.HasPrefix(stanzaFileOutput, "# --- start/stanza AUDIT"))
	assert.True(t, strings.Contains(stanzaFileOutput, "# --- start/stanza APPDYNAMICS"))
	assert.True(t, strings.Contains(stanzaFileOutput, "index = " + splunkAppDynamicsIndex))
	assert.True(t, strings.Contains(stanzaFileOutput, "index = " + splunkAuditIndex))
	t.Log(stanzaFileOutput)
}

func TestGenerateNoStanzas(t *testing.T) {
	dir, err := ioutil.TempDir("", "radishtest")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	outputFileName := dir
	os.Setenv("SPLUNK_INDEX", "")
	os.Setenv("SPLUNK_AUDIT_INDEX", "")
	os.Setenv("SPLUNK_APPDYNAMICS_INDEX", "")

	err = GenerateStanzas("", "", "", "", "", outputFileName)
	assert.NoError(t, err)
	_, err = os.Stat(outputFileName + "/application.splunk")
	assert.True(t, os.IsNotExist(err))
}

func generalStanzaFormat(stanzaFile string, entries int) bool {
	hostName := os.Getenv("HOSTNAME")
	podNamespace := os.Getenv("POD_NAMESPACE")
	appName := os.Getenv("APP_NAME")

	returnValue := true
	if strings.Count(stanzaFile, "# --- start/stanza") != entries {
		returnValue = false
	}
	if strings.Count(stanzaFile, "# --- end/stanza") != entries {
		returnValue = false
	}
	if strings.Count(stanzaFile, "disabled = false") != entries {
		returnValue = false
	}
	if strings.Count(stanzaFile, "followTail = 0") != entries {
		returnValue = false
	}
	if strings.Count(stanzaFile, fmt.Sprintf("_meta = environment::%s application::%s", podNamespace, appName)) != entries {
		returnValue = false
	}
	if strings.Count(stanzaFile, "[monitor://.") != entries {
		returnValue = false
	}
	if strings.Count(stanzaFile, "host = "+hostName) != entries {
		returnValue = false
	}
	return returnValue
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
