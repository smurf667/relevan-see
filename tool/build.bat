@echo off
set GOARCH=amd64

echo "Building for Linux..."
set GOOS=linux
go build -o bin/relevan-see

echo "Building for Windows..."
set GOOS=windows
go build -o bin/relevan-see.exe
