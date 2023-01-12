package tests

import (
	"bufio"
	"context"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/ptp-operator/test/pkg"
	"github.com/redhat-eets/sno-tests/test/pkg/client"
	"github.com/redhat-eets/sno-tests/test/pkg/consts"
	"github.com/redhat-eets/sno-tests/test/pkg/devices"
	"github.com/redhat-eets/sno-tests/test/pkg/pods"
)

var _ = Describe("PTP T-GM", func() {
	client := client.New("")

	Context("WPC GNSS verifications", func() {
		var ptpRunningPods []*corev1.Pod
		var ttyGNSS string
		var testPort string

		BeforeEach(func() {
			var err error

			testPort = devices.WPCDevicePort
			Expect(testPort).To(Not(BeEmpty()), "test port is not set")

			ptpRunningPods, err = pods.GetPTPDaemonPods(client)
			Expect(err).NotTo(HaveOccurred())

			vendor, device, tty, err := devices.GetDevInfo(client, testPort, ptpRunningPods[0])
			Expect(err).NotTo(HaveOccurred())
			ttyGNSS = tty
			if vendor != "0x8086" && device != "0x1593" {
				Skip("NIC device is not based on E810")
			}
		})

		It("Should check GNSS signal from the tty device in the host", func() {
			commands := []string{
				"/bin/sh", "-c", "timeout 2 cat " + ttyGNSS,
			}
			buf, _ := pods.ExecCommand(client, ptpRunningPods[0], pkg.PtpContainerName, commands)
			outstring := buf.String()
			Expect(outstring).To(Not(BeEmpty()))

			// These two are bad: http://aprs.gids.nl/nmea/#rmc
			// $GNRMC,,V,,,,,,,,,,N,V*37
			// $GNGGA,,,,,,0,00,99.99,,,,,,*56
			s := strings.Split(outstring, ",")
			Expect(len(s)).To(BeNumerically(">", 7), "Failed to parse GNSS string: %s", outstring)

			By("validating TTY GNSS GNRMC GPS/Transit data and GNGGA Positioning System Fix Data")
			if strings.Contains(s[0], "GNRMC") {
				Expect(s[2]).To(Not(Equal("V")))
			} else if strings.Contains(s[0], "GNGGA") {
				Expect(s[6]).To(Not(Equal("0")))
			}
		})

		It("Should check GNSS signal from PTP daemon log", func() {
			_, err := pods.GetLog(client, ptpRunningPods[0], pkg.PtpContainerName)
			Expect(err).NotTo(HaveOccurred(), "Error to find GNSS log in PTP daemon due to %s", err)
			result := pods.WaitUntilLogIsDetectedRegex(client, ptpRunningPods[0], pkg.TimeoutIn10Minutes, "nmea sentence: GNRMC(.*)")

			By("validating TTY GNSS GNRMC GPS/Transit data")
			s := strings.Split(result, ",")
			// Expecting: ,230304.00,A,4233.01530,N,07112.87856,W,0.002,,071222,,,A,V
			Expect(s[2]).To((Equal("A")))
		})

		It("Should check a valid 1PPS signal coming from the GNSS arrives the DPLL", func() {
			podDef, err := pods.DefinePodOnHost(consts.TestNamespace)
			Expect(err).NotTo(HaveOccurred())

			podOnHost, err := client.Pods(consts.TestNamespace).Create(context.Background(), podDef, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())

			err = pods.WaitForPhase(client, podOnHost, corev1.PodRunning, pkg.TimeoutIn3Minutes)
			Expect(err).NotTo(HaveOccurred())

			devCGU, err := devices.GetDevCGU(client, testPort, podOnHost)
			Expect(err).NotTo(HaveOccurred())

			PPSDPLL := strings.Split(devCGU, "PPS DPLL:")[1]

			PPSDPLLRef := ""
			PPSDPLLStatus := ""
			scanner := bufio.NewScanner(strings.NewReader(PPSDPLL))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "Current reference:") {
					PPSDPLLRef = strings.Split(line, ":")[1]
					PPSDPLLRef = strings.TrimSpace(PPSDPLLRef)
				}
				if strings.Contains(line, "Status:") {
					PPSDPLLStatus = strings.Split(line, ":")[1]
					PPSDPLLStatus = strings.TrimSpace(PPSDPLLStatus)
				}
			}

			Expect(PPSDPLLRef).To(Equal("GNSS-1PPS"), "failed to validate PPS DPLL current reference")
			Expect(PPSDPLLStatus).To(Equal("locked_ho_acq"), "failed to validate PPS DPLL status")
		})
	})
})
