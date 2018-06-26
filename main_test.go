package main

import (
	"os"
	"testing"

)

func TestMain(t *testing.T) {
	t.Log("testing TestMain")
	os.Args = []string{"bin/amd64/radish", "generateSplunkStanzas", "-c", "pkg/splunk/testdata/testconfig", "-o", "test123"}
	main()
}
