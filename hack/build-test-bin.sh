#!/bin/bash

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
	origin_url="https://github.com/redhat-eets/sno-tests/"
	git_tag=$(git describe --tags)
	if [ $? -ne 0 ]; then
		git_tag=$(git rev-parse HEAD)
		if [ $? -ne 0 ]; then
			origin_url="Not Found"
		else
			origin_url="${origin_url}tree/${git_tag}"
		fi
	else
		git_tag=$(echo $git_tag | sort -V | tail -1)
		origin_url="${origin_url}releases/tag/${git_tag}"
	fi
	ginkgo build -ldflags "-X 'github.com/redhat-eets/sno-tests/test/ptp.origin_url=${origin_url}'" ./test/"$suite"
	mv ./test/"$suite"/"$suite".test "$target"
}

build_and_move_suite "ptp" "./bin/ptptests"



