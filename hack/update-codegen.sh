#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# TODO: Remove once "Making unsupported type entry" in gengo is fixed
export GODEBUG="gotypesalias=0"

THIS_PKG="github.com/ironcore-dev/ironcore"
SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$SCRIPT_DIR/.."

VGOPATH="$VGOPATH"
MODELS_SCHEMA="$MODELS_SCHEMA"
OPENAPI_GEN="$OPENAPI_GEN"

VIRTUAL_GOPATH="$(mktemp -d)"
trap 'rm -rf "$VIRTUAL_GOPATH"' EXIT

# Setup virtual GOPATH so the codegen tools work as expected.
(cd "$PROJECT_ROOT"; go mod download && "$VGOPATH" -o "$VIRTUAL_GOPATH")

export GOROOT="${GOROOT:-"$(go env GOROOT)"}"
export GOPATH="$VIRTUAL_GOPATH"

CODE_GEN_DIR=$(go list -m -f '{{.Dir}}' k8s.io/code-generator)
source "${CODE_GEN_DIR}/kube_codegen.sh"

export TERM="xterm-256color"

bold="$(tput bold)"
blue="$(tput setaf 4)"
normal="$(tput sgr0)"


CLIENT_VERSION_GROUPS="core:v1alpha1 compute:v1alpha1 ipam:v1alpha1 networking:v1alpha1 storage:v1alpha1"
ALL_VERSION_GROUPS="common:v1alpha1 $CLIENT_VERSION_GROUPS"

echo "${bold}Public types${normal}"

echo "Generating ${blue}deepcopy, defaulter, and conversion${normal}"
kube::codegen::gen_helpers \
  --boilerplate "$SCRIPT_DIR/boilerplate.go.txt" \
  "$PROJECT_ROOT/api"

# NOTE: unsure why, but openapi-gen opens files not in read-only mode, so let's
#       workaround this for now by setting chmod for relevant modules
#       https://github.com/kubernetes/kubernetes/issues/136295
declare -a GOMODS=(
  "k8s.io/apimachinery"
  "k8s.io/api"
)
echo "Setting permissions for files of relevant go modules to 644"
for MOD in "${GOMODS[@]}"; do
  find "$(go list -json -m -u "${MOD}" | jq -r '.Dir')" -type f -exec chmod 644 -- {} +
done

echo "Generating ${blue}openapi${normal}"
kube::codegen::gen_openapi \
    --output-dir "${PROJECT_ROOT}/client-go/openapi" \
    --output-pkg "${THIS_PKG}/client-go/openapi" \
    --report-filename "$PROJECT_ROOT/client-go/openapi/api_violations.report" --update-report \
    --output-model-name-file "zz_generated.model_name.go" \
    --boilerplate "${PROJECT_ROOT}/hack/boilerplate.go.txt" \
    --extra-pkgs "k8s.io/api/core/v1" \
    "${PROJECT_ROOT}/api"

echo "Generating ${blue}client, lister, informer, and applyconfiguration${normal}"
applyconfigurationgen_external_apis+=("k8s.io/apimachinery/pkg/apis/meta/v1:k8s.io/client-go/applyconfigurations/meta/v1")
for GV in ${ALL_VERSION_GROUPS}; do
  IFS=: read -r G V <<<"${GV}"
  applyconfigurationgen_external_apis+=("${THIS_PKG}/api/${G}/${V}:${THIS_PKG}/client-go/applyconfigurations/${G}/${V}")
done
applyconfigurationgen_external_apis_csv=$(IFS=,; echo "${applyconfigurationgen_external_apis[*]}")

# Do not rely on process substitution / GNU bash
tmp_schema_file=$(mktemp)
trap 'rm -f "$tmp_schema_file"' EXIT
"$MODELS_SCHEMA" --openapi-package "${THIS_PKG}/client-go/openapi" --openapi-title "ironcore" > "$tmp_schema_file"

kube::codegen::gen_client \
  --with-applyconfig \
  --applyconfig-name "applyconfigurations" \
  --applyconfig-externals "${applyconfigurationgen_external_apis_csv}" \
  --applyconfig-openapi-schema "$tmp_schema_file" \
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
