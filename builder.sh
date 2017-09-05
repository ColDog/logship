#!/bin/sh

# Build modules.go.
cat > modules.go <<EOF
package main

import (
EOF
for plugin in $PLUGINS; do
  echo "	_ \"$plugin\"" >> modules.go
done
echo ")" >> modules.go

# Install plugins.
for plugin in $PLUGINS; do
  go get $plugin || exit 1
done

exec logship "$@"
