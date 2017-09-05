#!/bin/bash

plugin_list=($PLUGINS)

# Build modules.go.
cat > modules.go <<EOF
package main

import (
EOF
for plugin in "${plugin_list[@]}"; do
  echo "	_ \"$plugin\"" >> modules.go
done
echo ")" >> modules.go

go get -v ./... || exit 1
logship "$@"
