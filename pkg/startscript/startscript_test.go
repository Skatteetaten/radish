package startscript

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

var expectedStartScript = `
exec radish -Dfoo=bar -cp "/app/lib/metrics.jar:/app/lib/rt.jar:/app/lib/spring.jar" $JAVA_OPTS foo.bar.Main --logging.config=logback.xml
`

func TestGenerateStartscript(t *testing.T) {
	t.Log("test")
	json, err := ioutil.ReadFile("testconfig")
	assert.NoError(t, err)

	fmt.Println("say hi")
	t.Log("test")
	t.Logf("dat: %d", json)

	var data Data
	data, err = unMarshalJSON(json)
	assert.NoError(t, err)

	writer := newStartScript(data)
	buffer := new(bytes.Buffer)
	err = writer(buffer)
	assert.NoError(t, err)

	startscript := buffer.String()
	assert.Equal(t, expectedStartScript, startscript)

}
