#!/bin/bash

# Go to the script's directory
cd "$(dirname "$0")"

# Clean the bin/ directory
rm -rf bin/
mkdir -p bin

# Build the project
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -o bin/lunarium

echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -o bin/lunarium.exe
zip -r bin/lunarium-win.zip bin/lunarium.exe

# https://chatgpt.com/share/67bb2120-de00-800a-baa6-2d9822e6af3f
#echo "Building for macOS..."
#CGO_ENABLED=1 CC=x86_64-apple-darwin20-clang CXX=x86_64-apple-darwin20-clang++ GOOS=darwin GOARCH=amd64 go build -o bin/lunarium-mac
#CGO_ENABLED=1 CC=aarch64-apple-darwin20-clang CXX=aarch64-apple-darwin20-clang++ GOOS=darwin GOARCH=arm64 go build -o bin/lunarium-mac-arm


sudo cp bin/lunarium /usr/local/bin/lunarium

echo "Build complete: bin/lunarium"
