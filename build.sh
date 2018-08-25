#!/bin/bash

go build main.go
cp main bin/goproxy
mkdir output -p
cp bin config output -r
find output -name ".svn" | xargs rm -rf
cd output
tar czvf goproxy.tar.gz bin config
cd -
