#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
CODEGEN_PKG="${CODEGEN_PKG:-"$( (go mod download > /dev/null 2>&1) && go list -m -f '{{.Dir}}' k8s.io/code-generator)"}"

VGOPATH="$(mktemp -d)"
trap 'rm -rf "$VGOPATH"' EXIT

# Setup virtual GOPATH so the codegen tools work as expected.
(cd "$SCRIPT_DIR/.."; go run github.com/onmetal/vgopath "$VGOPATH")

export GOPATH="$VGOPATH"
export GO111MODULE=off

bash "$CODEGEN_PKG"/generate-internal-groups.sh \
  deepcopy,defaulter,conversion,client,lister,informer \
  github.com/onmetal/onmetal-api/generated \
  github.com/onmetal/onmetal-api/apis \
  github.com/onmetal/onmetal-api/apis \
  "compute:v1alpha1 storage:v1alpha1 networking:v1alpha1 ipam:v1alpha1" \
  --output-base "$VGOPATH/src" \
  --go-header-file "$SCRIPT_DIR/boilerplate.go.txt"

bash "$CODEGEN_PKG"/generate-groups.sh \
  deepcopy \
  github.com/onmetal/onmetal-api/generated \
  github.com/onmetal/onmetal-api/apis \
  github.com/onmetal/onmetal-api/apis \
  "common:v1alpha1" \
  --output-base "$VGOPATH/src" \
  --go-header-file "$SCRIPT_DIR/boilerplate.go.txt"

bash "$CODEGEN_PKG"/generate-internal-groups.sh \
  openapi \
  github.com/onmetal/onmetal-api/generated \
  github.com/onmetal/onmetal-api/apis \
  github.com/onmetal/onmetal-api/apis \
  "common:v1alpha1 compute:v1alpha1 storage:v1alpha1 networking:v1alpha1 ipam:v1alpha1" \
  --output-base "$VGOPATH/src" \
  --go-header-file "$SCRIPT_DIR/boilerplate.go.txt" \
  --input-dirs=k8s.io/api/core/v1 \
  --input-dirs=k8s.io/apimachinery/pkg/api/resource \
  --report-filename "$SCRIPT_DIR/../generated/openapi/api_violations.report"
