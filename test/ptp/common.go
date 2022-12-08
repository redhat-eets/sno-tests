package ptp_test

import (
	"strings"

	"github.com/openshift/ptp-operator/test/pkg"
	"github.com/redhat-eets/sno-tests/test/pkg/client"
	"github.com/redhat-eets/sno-tests/test/pkg/pods"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
)

func getDevInfo(api *client.ClientSet, intf string, ptpPod *corev1.Pod) (vendor, device, ttyGNSS string) {
	var commands []string
	var outstring string

	commands = []string{
		"readlink", "/sys/class/net/" + intf + "/device",
	}
	buf1, _ := pods.ExecCommand(api, ptpPod, pkg.PtpContainerName, commands)
	outstring = buf1.String()
	s := strings.Split(strings.TrimSpace(outstring), "/")
	busid := s[len(s)-1]
	logrus.Infof("busid: %s", busid)

	parts := strings.Split(busid, ":")
	ttyGNSS = parts[1] + strings.Split(parts[2], ".")[0]
	ttyGNSS = "/dev/ttyGNSS_" + ttyGNSS + "_0"
	commands = []string{
		"cat", "/sys/class/net/" + intf + "/device/device",
	}
	buf2, _ := pods.ExecCommand(api, ptpPod, pkg.PtpContainerName, commands)
	outstring = buf2.String()
	device = strings.TrimSpace(outstring)
	commands = []string{
		"cat", "/sys/class/net/" + intf + "/device/vendor",
	}
	buf3, _ := pods.ExecCommand(api, ptpPod, pkg.PtpContainerName, commands)
	outstring = buf3.String()
	vendor = strings.TrimSpace(outstring)
	logrus.Infof("vendor: %s, device ID: %s, ttyGNSS: %s", vendor, device, ttyGNSS)
	return
}
