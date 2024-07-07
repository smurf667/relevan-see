#!/bin/sh
export GOARCH=amd64

echo "Building for Linux..."
export GOOS=linux
go build -o bin/relevan-see

echo "Building for Windows..."
export GOOS=windows
go build -o bin/relevan-see.exe
