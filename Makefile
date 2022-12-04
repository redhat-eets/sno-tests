IMAGE_BUILD_CMD ?= "podman"

.PHONY: test-bin

deps-update:
	go mod tidy

test-bin:
	@echo "Making test binary"
	hack/build-test-bin.sh

