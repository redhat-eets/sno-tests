#!/bin/bash

set -e
set +x

if ! which go; then
  echo "No go command available"
  exit 1
fi

GOPATH="${GOPATH:-~/go}"
export GOFLAGS=${GOFLAGS:-"-mod=mod"}
export GO111MODULE="on"
export PATH=$PATH:$GOPATH/bin

if ! which ginkgo; then
	echo "Downloading ginkgo tool"
	go install github.com/onsi/ginkgo/ginkgo
fi

mkdir -p bin

function build_and_move_suite {
	suite=$1
	target=$2
	ginkgo build ./test/"$suite"
	mv ./test/"$suite"/"$suite".test "$target"
}

build_and_move_suite "ptp" "./bin/ptptests"



