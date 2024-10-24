#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

THIS_PKG="github.com/ironcore-dev/ironcore"
SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$SCRIPT_DIR/.."

CODE_GEN_DIR=$(go list -m -f '{{.Dir}}' k8s.io/code-generator)
source "${CODE_GEN_DIR}/kube_codegen.sh"

export TERM="xterm-256color"

bold="$(tput bold)"
blue="$(tput setaf 4)"
normal="$(tput sgr0)"

function qualify-gvs() {
  APIS_PKG="$1"
  GROUPS_WITH_VERSIONS="$2"
  join_char=""
  res=""

  for GVs in ${GROUPS_WITH_VERSIONS}; do
    IFS=: read -r G Vs <<<"${GVs}"

    for V in ${Vs//,/ }; do
      res="$res$join_char$APIS_PKG/$G/$V"
      join_char=" "
    done
  done

  echo "$res"
}

VGOPATH="$VGOPATH"
MODELS_SCHEMA="$MODELS_SCHEMA"
OPENAPI_GEN="$OPENAPI_GEN"

VIRTUAL_GOPATH="$(mktemp -d)"
trap 'rm -rf "$VIRTUAL_GOPATH"' EXIT

# Setup virtual GOPATH so the codegen tools work as expected.
(cd "$PROJECT_ROOT"; go mod download && "$VGOPATH" -o "$VIRTUAL_GOPATH")

export GOROOT="${GOROOT:-"$(go env GOROOT)"}"
export GOPATH="$VIRTUAL_GOPATH"

CLIENT_GROUPS="core compute ipam networking storage"
CLIENT_VERSION_GROUPS="core:v1alpha1 compute:v1alpha1 ipam:v1alpha1 networking:v1alpha1 storage:v1alpha1"
ALL_VERSION_GROUPS="common:v1alpha1 $CLIENT_VERSION_GROUPS"

echo "${bold}Public types${normal}"

echo "Generating ${blue}deepcopy, defaulter, and conversion${normal}"
kube::codegen::gen_helpers \
  --boilerplate "$SCRIPT_DIR/boilerplate.go.txt" \
  "$PROJECT_ROOT/api"

echo "Generating ${blue}openapi${normal}"
input_dirs=($(qualify-gvs "${THIS_PKG}/api" "$ALL_VERSION_GROUPS"))
"$OPENAPI_GEN" \
  --output-dir "$PROJECT_ROOT/client-go/openapi" \
  --output-pkg "${THIS_PKG}/client-go/openapi" \
  --output-file "zz_generated.openapi.go" \
  --report-filename "$PROJECT_ROOT/client-go/openapi/api_violations.report" \
  --go-header-file "$SCRIPT_DIR/boilerplate.go.txt" \
  "k8s.io/apimachinery/pkg/apis/meta/v1" \
  "k8s.io/apimachinery/pkg/runtime" \
  "k8s.io/apimachinery/pkg/version" \
  "k8s.io/api/core/v1" \
  "k8s.io/apimachinery/pkg/api/resource" \
  "${input_dirs[@]}"

echo "Generating ${blue}client, lister, informer, and applyconfiguration${normal}"
applyconfigurationgen_external_apis+=("k8s.io/apimachinery/pkg/apis/meta/v1:k8s.io/client-go/applyconfigurations/meta/v1")
for GV in ${ALL_VERSION_GROUPS}; do
  IFS=: read -r G V <<<"${GV}"
  applyconfigurationgen_external_apis+=("${THIS_PKG}/api/${G}/${V}:${THIS_PKG}/client-go/applyconfigurations/${G}/${V}")
done
applyconfigurationgen_external_apis_csv=$(IFS=,; echo "${applyconfigurationgen_external_apis[*]}")
kube::codegen::gen_client \
  --with-applyconfig \
  --applyconfig-name "applyconfigurations" \
  --applyconfig-externals "${applyconfigurationgen_external_apis_csv}" \
  --applyconfig-openapi-schema <("$MODELS_SCHEMA" --openapi-package "${THIS_PKG}/client-go/openapi" --openapi-title "ironcore") \
  --clientset-name "ironcore" \
  --listers-name "listers" \
  --informers-name "informers" \
  --with-watch \
  --output-dir "$PROJECT_ROOT/client-go" \
  --output-pkg "${THIS_PKG}/client-go" \
  --boilerplate "$SCRIPT_DIR/boilerplate.go.txt" \
  "$PROJECT_ROOT/api"

echo "${bold}Internal types${normal}"

echo "Generating ${blue}deepcopy, defaulter, and conversion${normal}"
kube::codegen::gen_helpers \
  --boilerplate "$SCRIPT_DIR/boilerplate.go.txt" \
  "$PROJECT_ROOT/internal/apis"
