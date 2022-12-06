#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

FILE="$1"
EXPRESSION="$2"

sed "$EXPRESSION" "$FILE" > "$FILE.bak"
mv "$FILE.bak" "$FILE"
