package ptp_test

import (
	"context"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/ptp-operator/test/pkg"
	"github.com/redhat-eets/sno-tests/test/pkg/pods"
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
	})

	Context("WPC GNSS verifications", func() {
		var ptpRunningPods []*corev1.Pod
		var ttyGNSS string
		BeforeEach(func() {
			// TODO test all ports
			if topo.PTP.GM.PortTester == nil {
				Skip("No T-GM port not specified")
			}
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
			vendor, device, tty := getDevInfo(clients["GM"], *topo.PTP.GM.PortTester, ptpRunningPods[0])
			ttyGNSS = tty
			if vendor != "0x8086" && device != "0x1593" {
				Skip("NIC is not a WPC")
			}
		})

		It("Should check GNSS signal on the host", func() {
			commands := []string{
				"/bin/sh", "-c", "timeout 2 cat " + ttyGNSS,
			}
			buf, _ := pods.ExecCommand(clients["GM"], ptpRunningPods[0], pkg.PtpContainerName, commands)
			outstring := buf.String()
			Expect(outstring).To(Not(BeEmpty()))
			logrus.Infof("captured log: %s", outstring)
			// These two are bad: http://aprs.gids.nl/nmea/#rmc
			// $GNRMC,,V,,,,,,,,,,N,V*37
			// $GNGGA,,,,,,0,00,99.99,,,,,,*56
			s := strings.Split(outstring, ",")
			Expect(len(s)).To(BeNumerically(">", 7), "Failed to parse GNSS string: %s", outstring)
			if strings.Contains(s[0], "GNRMC") {
				Expect(s[2]).To(Not(Equal("V")))
			} else if strings.Contains(s[0], "GNGGA") {
				Expect(s[6]).To(Not(Equal("0")))
			}
		})

		It("Should check GNSS from PTP log", func() {
			if topo.PTP.GM.PortTester == nil {
				Skip("WPC port not provided")
			}
			_, err := pods.GetLog(clients["GM"], ptpRunningPods[0], pkg.PtpContainerName)
			Expect(err).NotTo(HaveOccurred(), "Error to find needed log due to %s", err)
			result := pods.WaitUntilLogIsDetectedRegex(clients["GM"], ptpRunningPods[0], pkg.Timeout10Seconds, "nmea sentence: GNRMC(.*)")
			logrus.Infof("captured log: %s", result)
			s := strings.Split(result, ",")
			// Expecting: ,230304.00,A,4233.01530,N,07112.87856,W,0.002,,071222,,,A,V
			Expect(s[2]).To((Equal("A")))

		})

	})

	Context("Verifications from Tester", func() {
		var ptpRunningPods []*corev1.Pod
		BeforeEach(func() {
			if topo.PTP.Tester == nil {
				Skip("Tester setting not provided")
			}
			ptpPods, err := clients["Tester"].CoreV1().Pods(pkg.PtpLinuxDaemonNamespace).List(context.Background(), metav1.ListOptions{LabelSelector: "app=linuxptp-daemon"})
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

		It("Tester should be able to sync", func() {
			_, err := pods.GetLog(clients["Tester"], ptpRunningPods[0], pkg.PtpContainerName)
			Expect(err).NotTo(HaveOccurred(), "Error to find needed log due to %s", err)
			result := pods.WaitUntilLogIsDetectedRegex(clients["Tester"], ptpRunningPods[0], pkg.Timeout10Seconds, "ptp4l.*rms (.*)")
			logrus.Infof("captured log: %s", result)
		})

	})

})
