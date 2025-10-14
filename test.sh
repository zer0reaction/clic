#!/bin/bash

if [[ $# -ne 1 ]]; then
    echo "You must specify the name of the test file"
    exit 1
fi

go run ./cmd/main.go -bf "-ggdb -o /tmp/out ./extern.c" $1 && /tmp/out
