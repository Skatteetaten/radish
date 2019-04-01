package main

import (
	"os"
	"testing"
)

func TestMain(t *testing.T) {
	t.Log("testing TestMain")
	os.Setenv("POD_NAMESPACE", "1")
	os.Setenv("APP_NAME", "appname")
	os.Setenv("HOSTNAME", "host")
	os.Setenv("SPLUNK_INDEX", "myindex")
	os.Args = []string{"bin/amd64/radish", "generateSplunkStanzas", "-o", "test123"}
	main()
}
