#!/bin/bash

parent_path=$(pwd)
find . -type d | while read dir; do
  if [[ -f "$dir/.golangci.yml" ]]; then
    cd $dir; go mod tidy; golangci-lint run ;cd $parent_path
  fi
done