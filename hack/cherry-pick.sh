#!/usr/bin/env bash

set -euo pipefail

usage() {
    echo "Usage: $0 <PR-number> <release-branch>"
    echo ""
    echo "Cherry-pick a merged PR into a release branch and open a new PR."
    echo ""
    echo "Examples:"
    echo "  $0 123 release-v0.1"
    echo "  $0 456 release-v0.2"
    echo ""
    echo "Prerequisites:"
    echo "  - gh CLI installed and authenticated"
    echo "  - Clean working tree"
    exit 1
}

if [[ $# -ne 2 ]]; then
    usage
fi

PR_NUMBER="$1"
RELEASE_BRANCH="$2"

# Verify gh CLI is available
if ! command -v gh &>/dev/null; then
    echo "Error: gh CLI is not installed. See https://cli.github.com/"
    exit 1
fi

# Verify clean working tree
if [[ -n "$(git status --porcelain)" ]]; then
    echo "Error: working tree is not clean. Please commit or stash your changes."
    exit 1
fi

# Verify the release branch exists on remote
if ! git ls-remote --exit-code --heads origin "${RELEASE_BRANCH}" &>/dev/null; then
    echo "Error: branch '${RELEASE_BRANCH}' does not exist on remote."
    exit 1
fi

# Get the merge commit SHA for the PR
MERGE_COMMIT=$(gh pr view "${PR_NUMBER}" --json mergeCommit --jq '.mergeCommit.oid')
if [[ -z "${MERGE_COMMIT}" ]]; then
    echo "Error: PR #${PR_NUMBER} has no merge commit. Is it merged?"
    exit 1
fi

PR_TITLE=$(gh pr view "${PR_NUMBER}" --json title --jq '.title')
CHERRY_PICK_BRANCH="cherry-pick-${PR_NUMBER}-into-${RELEASE_BRANCH}"

echo "Cherry-picking PR #${PR_NUMBER} (\"${PR_TITLE}\") into ${RELEASE_BRANCH}"
echo "  Merge commit: ${MERGE_COMMIT}"
echo "  Branch: ${CHERRY_PICK_BRANCH}"
echo ""

# Fetch latest remote state
git fetch origin "${RELEASE_BRANCH}"

# Create cherry-pick branch from the release branch
git switch -c "${CHERRY_PICK_BRANCH}" "origin/${RELEASE_BRANCH}"

# Cherry-pick the merge commit using the first parent
if ! git cherry-pick -x -m1 "${MERGE_COMMIT}"; then
    echo ""
    echo "Cherry-pick has conflicts. Please resolve them, then run:"
    echo "  git cherry-pick --continue"
    echo "  git push origin ${CHERRY_PICK_BRANCH}"
    echo "  gh pr create --base ${RELEASE_BRANCH} --title \"🍒 [${RELEASE_BRANCH}] ${PR_TITLE}\" --body \"Cherry-pick of #${PR_NUMBER} into ${RELEASE_BRANCH}.\""
    exit 1
fi

# Push and create PR
git push origin "${CHERRY_PICK_BRANCH}"
gh pr create \
    --base "${RELEASE_BRANCH}" \
    --title "🍒 [${RELEASE_BRANCH}] ${PR_TITLE}" \
    --body "Cherry-pick of #${PR_NUMBER} into \`${RELEASE_BRANCH}\`."

echo ""
echo "Done."
