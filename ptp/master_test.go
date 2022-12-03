package ptp_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Master", func() {
	BeforeEach(func() {
		if topo.PTP.GM == nil {
			Skip("T-GM setting not provided")
		}
		if topo.PTP.GM.Node == "" {
			Skip("T-GM node not provided")
		}

	})
	It("reports the env var", func() {
		Expect(os.Getenv("SHELL")).To(Equal("/bin/bash"))
	})

})
