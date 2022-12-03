package ptp_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/ptp-operator/test/pkg"
	"github.com/redhat-eets/sno-tests/pkg/pods"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("T-GM", func() {

	BeforeEach(func() {
		if topo.PTP.GM == nil {
			Skip("T-GM setting not provided")
		}
		if topo.PTP.GM.Node == "" {
			Skip("T-GM node not provided")
		}
		Expect(clients["GM"].Client).NotTo(BeNil())
	})

	Context("WPC GNSS verifications", func() {
		var ptpRunningPods []*corev1.Pod
		BeforeEach(func() {
			ptpPods, err := clients["GM"].CoreV1().Pods(pkg.PtpLinuxDaemonNamespace).List(context.Background(), metav1.ListOptions{LabelSelector: "app=linuxptp-daemon"})
			Expect(err).NotTo(HaveOccurred())
			Expect(len(ptpPods.Items)).To(BeNumerically(">", 0), "linuxptp-daemon is not deployed on cluster")
			logrus.Infof("number of ptpPods: %d", len(ptpPods.Items))
			ptpRunningPods = []*corev1.Pod{}
			for podIndex := range ptpPods.Items {
				ptpRunningPods = append(ptpRunningPods, &ptpPods.Items[podIndex])
			}

			Expect(ptpRunningPods[0]).NotTo(BeNil())
			logrus.Infof("number of ptpRunningPods: %d", len(ptpRunningPods))
		})

		It("Should check GNSS signal", func() {
			if topo.PTP.GM.PortTester == nil {
				Skip("WPC port not provided")
			}
			_, err := pods.GetLog(clients["GM"], ptpRunningPods[0], pkg.PtpContainerName)
			Expect(err).NotTo(HaveOccurred(), "Error to find needed log due to %s", err)
			result := pods.WaitUntilLogIsDetectedRegex(clients["GM"], ptpRunningPods[0], pkg.Timeout10Seconds, "nmea sentence: GNRMC(.*)")
			logrus.Infof("captured log: %s", result)

		})

	})

})
