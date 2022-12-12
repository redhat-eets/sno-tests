package ptp_test

import (
	"strings"

	"github.com/openshift/ptp-operator/test/pkg"
	"github.com/redhat-eets/sno-tests/test/pkg/client"
	"github.com/redhat-eets/sno-tests/test/pkg/pods"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"github.com/go-git/go-git/v5"
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

func getOriginUrl() (origin_url string) {
	var replacer = strings.NewReplacer("git@", "https://", ":", "/", ".git", "/tree/")
	r, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{
                        DetectDotGit: true,
        })
        if err == nil {
                ref_head, err_head := r.Head()
                ref_origin, err_origin := r.Remotes()
		if err_origin == nil {
			origin_url = replacer.Replace(strings.Fields(ref_origin[0].String())[1])
			if err_head == nil {
				commit_hash := strings.Fields(ref_head.String())[0]
				origin_url = origin_url + commit_hash
			}
                }
        }

	if origin_url == "" {
		origin_url = "Not Found"
	}

	return
}
