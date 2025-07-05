#!/bin/bash

echo "Running Architecture Tests..."

if ! command -v arch-go > /dev/null; then
    echo "Installing arch-go..."
    go install github.com/arch-go/arch-go@latest
fi

if [ ! -f "arch-go.yml" ]; then
    echo "❌ Error: arch-go.yml configuration file not found!"
    echo "Please make sure you're running this script from the project root directory."
    exit 1
fi

echo "Validating architecture with arch-go..."
arch-go

echo "Architecture tests completed!"

echo "Press any key to continue..."
read -n 1 -s