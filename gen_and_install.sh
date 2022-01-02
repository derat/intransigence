#!/bin/sh -e

# Regenerate the files under render/ and compile and install the executable.
# This is convenient for development.

cd render && go generate
cd .. && go install ./cmd/...
