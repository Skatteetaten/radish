#!/bin/bash

set -e

make clean
make
make test-xml
make test-coverage

cd bin
tar -cvzf radish.tar.gz radish
