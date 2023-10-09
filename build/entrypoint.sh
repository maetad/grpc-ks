#!/bin/bash

set -e

if [ -z "$SSH_AUTH_SOCK" ]; then
  echo "No ssh agent detected"
else
  echo $SSH_AUTH_SOCK
  ssh-add -l
fi


export GO111MODULE=on
export GONOPROXY="github.com/maetad"
export GONOSUMDB="github.com/maetad"
export GOPRIVATE="github.com/maetad"

go mod tidy
