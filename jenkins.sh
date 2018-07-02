#!/bin/bash


type dep 2> /dev/null || /bin/sh -c "export GOPATH=$GOROOT && curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh"
type go-junit-report 2> /dev/null || go get -u github.com/jstemmer/go-junit-report
type gocov 2> /dev/null || go get github.com/axw/gocov/gocov
type gocov-xml 2> /dev/null || go get github.com/AlekSi/gocov-xml
type go-bindata 2> /dev/null || go get -u github.com/jteeuwen/go-bindata/...

make build-dirs

export GOPATH="$(pwd)/.go" 
# We need to run dep ensure from inside a **/src folder that is inside GOPATH
export JENKINSCODEPATH="src/github.com/skatteetaten/radish" 
echo ${GOPATH}/${JENKINSCODEPATH} 
cd ${GOPATH}/${JENKINSCODEPATH}

echo "RUNNING DEP ENSURE from $(pwd)"
dep ensure


export JUNIT_REPORT=TEST-junit.xml
export COBERTURA_REPORT=coverage.xml

# Go get is not the best way of installing.... :/
export PATH=$PATH:$HOME/go/bin

make clean 

#Create executable in /bin/amd64
make

#Run test and coverage
make test

cd bin/amd64
tar -cvzf radish.tar.gz radish
