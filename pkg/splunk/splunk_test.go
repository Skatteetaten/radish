package splunk

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateStanzas(t *testing.T) {
	outputFileName := "testStanza"
	GenerateStanzas("", helperGetTestPath("testconfig"), outputFileName)

	if _, err := os.Stat(outputFileName); err == nil {
		t.Logf("%s exists!", outputFileName)

		//clean up after test
		t.Logf("Deleting %s", outputFileName)
		err := os.Remove(outputFileName)
		if err != nil {
			fmt.Println(err)
			t.Log(err)
			return
		}

	}

}

func TestReadStanzasTemplate(t *testing.T) {
	stanzaFromBinData, err := readStanzasTemplate("")
	assert.NoError(t, err)

	stanzaString := string(stanzaFromBinData)
	t.Log(stanzaString)

	assert.True(t, strings.HasPrefix(stanzaString, "# --- start/stanza STDOUT"))

	stanzaFromBinData2, err := readStanzasTemplate("resources/default_stanzas_template")
	assert.NoError(t, err)

	stanzaString2 := string(stanzaFromBinData2)
	t.Log(stanzaString2)

	assert.True(t, strings.HasPrefix(stanzaString2, "# --- start/stanza STDOUT"))
}

func TestReadConfig(t *testing.T) {
	config, err := readConfigFile(helperGetTestPath("testconfig"))
	assert.NoError(t, err)
	assert.True(t, config.AppName == "appName")
}

//TODO move generic helper funcs to utils package or something
func helperGetTestPath(name string) string {
	path := filepath.Join("testdata", name) // relative path
	return path
}

func helperLoadBytes(t *testing.T, name string) []byte {
	path := helperGetTestPath(name)
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}
