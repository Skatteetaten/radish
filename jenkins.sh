#!/bin/bash

make clean
make
make test-xml
make test-coverage
go test -short -coverprofile=bin/cov.out `go list ./... | grep -v vendor/`

cd bin
tar -cvzf radish.tar.gz radish
