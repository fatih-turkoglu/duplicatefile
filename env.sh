#!/usr/bin/env bash
# Source in Bash or zsh: source ./env.sh
if [ -n "${ZSH_VERSION:-}" ]; then
  DUPLICATEFILE_PROJECT_ROOT="$(cd -- "$(dirname -- "${(%):-%x}")" && pwd)"
else
  DUPLICATEFILE_PROJECT_ROOT="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
fi
export GOPATH="${GOPATH:-$DUPLICATEFILE_PROJECT_ROOT/.tools/gopath}"
export GOCACHE="${GOCACHE:-$DUPLICATEFILE_PROJECT_ROOT/.tools/go-cache}"
export npm_config_cache="${npm_config_cache:-$DUPLICATEFILE_PROJECT_ROOT/.tools/npm-cache}"
export PATH="$DUPLICATEFILE_PROJECT_ROOT/.tools/bin:$GOPATH/bin:$DUPLICATEFILE_PROJECT_ROOT/.tools/go/bin:$PATH"
export DUPLICATEFILE_LANG_DIR="${DUPLICATEFILE_LANG_DIR:-$DUPLICATEFILE_PROJECT_ROOT/lang}"
