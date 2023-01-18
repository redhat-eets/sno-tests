package tests

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"

	"github.com/openshift/ptp-operator/test/pkg"
	"github.com/redhat-eets/sno-tests/test/pkg/client"
	"github.com/redhat-eets/sno-tests/test/pkg/consts"
	"github.com/redhat-eets/sno-tests/test/pkg/devices"
	"github.com/redhat-eets/sno-tests/test/pkg/pods"
)

var _ = Describe("PTP T-GM Validation", func() {
	client := client.New("")

	Context("Configuration Check", func() {
		var ptpRunningPod *v1.Pod
		var testPort string

		BeforeEach(func() {
			var err error

			testPort = devices.WPCDevicePort
			Expect(testPort).To(Not(BeEmpty()), "test port is not set")

			ptpRunningPod, err = pods.GetPTPDaemonPod(client)
			Expect(err).NotTo(HaveOccurred())

			vendor, device, _, err := devices.GetDevInfo(client, testPort, ptpRunningPod)
			Expect(err).NotTo(HaveOccurred())
			if vendor != "0x8086" && device != "0x1593" {
				Skip("NIC device is not based on E810")
			}
		})

		It("Should have the desired firmware version", func() {
			commands := []string{
				"/bin/sh", "-c", "ethtool -i " + testPort,
			}

			buf, err := pods.ExecCommand(client, ptpRunningPod, pkg.PtpContainerName, commands)
			outstring := buf.String()
			Expect(err).NotTo(HaveOccurred(), "Error to find device driver info due to: %s %s", err, outstring)

			By(fmt.Sprintf("checking the firmware version equals or greater than %.2f", consts.ICEDriverFirmwareVerMinVersion))
			scanner := bufio.NewScanner(strings.NewReader(outstring))
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "firmware-version:") {
					fullFirmwareVer := strings.Split(line, ":")[1]
					By(fmt.Sprintf("NVM firmware version installed: %s", fullFirmwareVer))
					firmwareVerNum := strings.Split(fullFirmwareVer, " ")[1]
					Expect(strconv.ParseFloat(firmwareVerNum, 64)).To(BeNumerically(">=", consts.ICEDriverFirmwareVerMinVersion), "linuxptp-daemon is not deployed on cluster")
				}
			}
		})
	})
})
