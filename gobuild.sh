#!/bin/bash

cd /go/src/github.com/ndslabs-irods-federate
go get
GOOS=linux GOARCH=amd64 go build -o build/bin/ndslabs-irods-federate-linux-amd64
