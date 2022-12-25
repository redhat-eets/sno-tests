//go:build multisnotgmtests
// +build multisnotgmtests

package multisno

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	_ "github.com/redhat-eets/sno-tests/test/ptp/tgm/multisno/tests"
)

func TestTGM(t *testing.T) {
	RegisterFailHandler(Fail)
}
