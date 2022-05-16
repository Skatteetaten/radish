#!/bin/bash

echo "Building version $VERSION..."

make clean
make test
make test-coverage
go test -short -coverprofile=bin/cov.out `go list ./... | grep -v vendor/`
make

cd bin
tar -cvzf radish.tar.gz radish
