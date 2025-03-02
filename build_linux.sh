#!/bin/bash

# Set build variables
APP_NAME="thelistbot"
MAIN_FILE="./cmd/theList.go"
OUTPUT_DIR="./bin"

# Create output directory if it doesn't exist
mkdir -p $OUTPUT_DIR

# Build for Linux AMD64 (most common)
echo "Building for Linux (64-bit)..."
GOOS=linux GOARCH=amd64 go build -o "$OUTPUT_DIR/${APP_NAME}-linux-amd64" $MAIN_FILE

# Build for Linux ARM64 (for Raspberry Pi 4, etc.)
echo "Building for Linux ARM64..."
GOOS=linux GOARCH=arm64 go build -o "$OUTPUT_DIR/${APP_NAME}-linux-arm64" $MAIN_FILE

# Build for Linux ARM (for older Raspberry Pi models)
echo "Building for Linux ARM..."
GOOS=linux GOARCH=arm go build -o "$OUTPUT_DIR/${APP_NAME}-linux-arm" $MAIN_FILE

echo "Build complete! Files are in the $OUTPUT_DIR directory."
ls -la $OUTPUT_DIR
