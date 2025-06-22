#!/bin/bash

echo "Running Unit Tests..."
go test -v ./internal/core/service/... -tags=unit

echo "Press any key to continue..."
read -n 1 -s