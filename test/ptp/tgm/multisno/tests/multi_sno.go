package tests

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"github.com/openshift/ptp-operator/test/pkg"
	"github.com/redhat-eets/sno-tests/test/pkg/client"
	"github.com/redhat-eets/sno-tests/test/pkg/pods"
)

var _ = Describe("PTP T-GM", func() {
	client := client.New("")

	Context("WPC GNSS verifications from Tester", func() {
		var ptpRunningPods []*corev1.Pod

		BeforeEach(func() {
			var err error
			ptpRunningPods, err = pods.GetPTPDaemonPods(client)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Tester should be able to sync", func() {
			_, err := pods.GetLog(client, ptpRunningPods[0], pkg.PtpContainerName)
			Expect(err).NotTo(HaveOccurred(), "Error to find needed log due to %s", err)
			result := pods.WaitUntilLogIsDetectedRegex(client, ptpRunningPods[0], pkg.Timeout10Seconds, "ptp4l.*rms (.*)")
			logrus.Infof("captured log: %s", result)
		})
	})
})
