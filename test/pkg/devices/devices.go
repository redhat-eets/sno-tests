package devices

import (
	"bufio"
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/ptp-operator/test/pkg"
	"github.com/redhat-eets/sno-tests/test/pkg/client"
	"github.com/redhat-eets/sno-tests/test/pkg/consts"
	"github.com/redhat-eets/sno-tests/test/pkg/pods"
	"github.com/sirupsen/logrus"
)

var WPCDevicePort string

// GetDevInfo returns the vendor, device and ttyGNSS of an interface
func GetDevInfo(client *client.ClientSet, intf string, ptpPod *corev1.Pod) (string, string, string, error) {
	busID, err := GetBusID(client, intf, ptpPod)
	if err != nil {
		return "", "", "", err
	}
	logrus.Infof("busID: %s", busID)

	parts := strings.Split(busID, ":")
	ttyGNSS := parts[1] + strings.Split(parts[2], ".")[0]
	ttyGNSS = "/dev/ttyGNSS_" + ttyGNSS
	commands := []string{
		"cat", "/sys/class/net/" + intf + "/device/device",
	}
	buf, err := pods.ExecCommand(client, ptpPod, pkg.PtpContainerName, commands)
	outstring := buf.String()
	if err != nil {
		return "", "", "", fmt.Errorf("error to get device info due to: %s %s", err, outstring)
	}

	device := strings.TrimSpace(outstring)

	commands = []string{
		"cat", "/sys/class/net/" + intf + "/device/vendor",
	}
	buf, err = pods.ExecCommand(client, ptpPod, pkg.PtpContainerName, commands)
	outstring = buf.String()
	if err != nil {
		return "", "", "", fmt.Errorf("error to get vendor due to: %s %s", err, outstring)
	}

	vendor := strings.TrimSpace(outstring)
	logrus.Infof("vendor: %s, device ID: %s, ttyGNSS: %s", vendor, device, ttyGNSS)

	return vendor, device, ttyGNSS, nil
}

// GetBusID returns the Bus ID of an interface
func GetBusID(client *client.ClientSet, intf string, ptpPod *corev1.Pod) (string, error) {
	commands := []string{
		"readlink", "/sys/class/net/" + intf + "/device",
	}
	buf, err := pods.ExecCommand(client, ptpPod, pkg.PtpContainerName, commands)
	outstring := buf.String()
	if err != nil {
		return "", fmt.Errorf("error to get busID due to: %s %s", err, outstring)
	}

	s := strings.Split(strings.TrimSpace(outstring), "/")
	busID := s[len(s)-1]

	return busID, nil
}

// InitWPCDevicePort sets the WPCDevicePort variable to one of the ports connected to WPC NIC.
// Needs a pod with a host mount to get the inforamtion.
func InitWPCDevicePort(client *client.ClientSet, hostPod *corev1.Pod) error {
	nodeDevicesList, err := client.NodePtpDevices(pkg.PtpLinuxDaemonNamespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	if len(nodeDevicesList.Items) == 0 {
		return fmt.Errorf("no NodePtpDevices found")
	}

	WPCDevicePort = ""
	for _, nodeDevice := range nodeDevicesList.Items {
		if WPCDevicePort != "" {
			break
		}

		for _, dev := range nodeDevice.Status.Devices {
			if WPCDevicePort != "" {
				break
			}

			devType, err := getDevType(client, dev.Name, hostPod)
			if err != nil {
				return err
			}

			if strings.Contains(devType, consts.WPCDeviceType) {
				WPCDevicePort = dev.Name
				break
			}
		}
	}

	return nil
}

// GetDevDPLLInfo returns the device DPLL info for an interface.
// Needs a pod with a host mount to get the inforamtion.
func GetDevDPLLInfo(client *client.ClientSet, intf string, hostPod *corev1.Pod) (string, string, error) {
	commands := []string{
		"chroot", "/host", "cat", "/sys/class/net/" + intf + "/device/dpll_1_state",
	}

	buf, err := pods.ExecCommand(client, hostPod, hostPod.Spec.Containers[0].Name, commands)
	outstring := buf.String()
	if err != nil {
		return "", "", fmt.Errorf("failed to get DPLL state for device %s due to: %s %s", intf, err, outstring)
	}
	state := strings.TrimSpace(outstring)

	commands = []string{
		"chroot", "/host", "cat", "/sys/class/net/" + intf + "/device/dpll_1_offset",
	}

	buf, err = pods.ExecCommand(client, hostPod, hostPod.Spec.Containers[0].Name, commands)
	outstring = buf.String()
	if err != nil {
		return "", "", fmt.Errorf("failed to get DPLL offset for device %s due to: %s %s", intf, err, outstring)
	}
	offset := strings.TrimSpace(outstring)

	return state, offset, nil
}

// getDevType returns the device type of an interface.
// Needs a pod with a host mount to get the inforamtion.
func getDevType(client *client.ClientSet, intf string, hostPod *corev1.Pod) (string, error) {
	command := []string{
		"chroot", "/host", "readlink", "/sys/class/net/" + intf + "/device",
	}
	buf, err := pods.ExecCommand(client, hostPod, hostPod.Spec.Containers[0].Name, command)
	outstring := buf.String()
	if err != nil {
		return "", fmt.Errorf("error to get busID due to: %s %s", err, outstring)
	}

	s := strings.Split(strings.TrimSpace(outstring), "/")
	busID := s[len(s)-1]

	commands := []string{
		"chroot", "/host", "lspci", "-v", "-nn", "-mm", "-s", busID,
	}
	buf, err = pods.ExecCommand(client, hostPod, hostPod.Spec.Containers[0].Name, commands)
	outstring = buf.String()
	if err != nil {
		return "", fmt.Errorf("lspci for %s failed due to: %s %s", busID, err, outstring)
	}

	scanner := bufio.NewScanner(strings.NewReader(outstring))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "SDevice:") {
			return strings.Split(line, ":")[1], nil
		}
	}

	return "", fmt.Errorf("failed to get device type for %s failed due to: %s %s", intf, err, outstring)
}
