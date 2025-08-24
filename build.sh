#!/usr/bin/bash

set -e

cd ../hypersonic-ui
bun run build
cp dist/ ../hypersonic/internal/routes/ -r
cd ../hypersonic
go build -o hypersonic cmd/hypersonic/main.go
