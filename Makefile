IMAGE_BUILD_CMD ?= "podman"
TESTS_REPORTS_PATH ?= /tmp/logs

.PHONY: test-bin

deps-update:
	go mod tidy

test-bin:
	@echo "Making test binary"
	hack/build-test-bin.sh

test-tgm-validation-only:
	rm -rf ${TESTS_REPORTS_PATH}/test_tgm_validation_logs
	mkdir -p ${TESTS_REPORTS_PATH}/test_tgm_validation_logs
	go test --tags=tgmvalidationtests -v ./test/ptp/tgm/validation -ginkgo.v -junit ${TESTS_REPORTS_PATH}/test_tgm_validation_logs -report ${TESTS_REPORTS_PATH}/test_tgm_validation_logs

test-tgm:
	rm -rf ${TESTS_REPORTS_PATH}/test_tgm_logs
	mkdir -p ${TESTS_REPORTS_PATH}/test_tgm_logs
	go test --tags=tgmvalidationtests -v ./test/ptp/tgm/validation -ginkgo.v -junit ${TESTS_REPORTS_PATH}/test_tgm_logs -report ${TESTS_REPORTS_PATH}/test_tgm_logs
	go test --tags=tgmfunctionaltests -v ./test/ptp/tgm/functional -ginkgo.v -junit ${TESTS_REPORTS_PATH}/test_tgm_logs -report ${TESTS_REPORTS_PATH}/test_tgm_logs

test-tgm-multisno:
	rm -rf ${TESTS_REPORTS_PATH}/test_tgm_multisno_logs
	mkdir -p ${TESTS_REPORTS_PATH}/test_tgm_multisno_logs
	go test --tags=multisnotgmtests -v ./test/ptp/tgm/multisno -ginkgo.v -junit ${TESTS_REPORTS_PATH}/test_tgm_multisno_logs -report ${TESTS_REPORTS_PATH}/test_tgm_multisno_logs
