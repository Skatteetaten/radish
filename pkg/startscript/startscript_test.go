package startscript

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

var expectedStartScript = `
exec radish -Dfoo=bar -cp "/app/lib/metrics.jar:/app/lib/rt.jar:/app/lib/spring.jar" $JAVA_OPTS foo.bar.Main --logging.config=logback.xml
`

func TestGenerateStartscript(t *testing.T) {
	t.Log("test")
	json := helperLoadBytes(t, "testconfig")

	data, err := unMarshalJSON(json)
	assert.NoError(t, err)

	writer := newStartScript(data)
	buffer := new(bytes.Buffer)
	err = writer(buffer)
	assert.NoError(t, err)

	startscript := buffer.String()
	assert.Equal(t, expectedStartScript, startscript)

}

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
