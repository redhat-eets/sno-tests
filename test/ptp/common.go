package ptp_test

import (
	"strings"
	"github.com/openshift/ptp-operator/test/pkg"
	"github.com/redhat-eets/sno-tests/test/pkg/client"
	"github.com/redhat-eets/sno-tests/test/pkg/pods"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"github.com/go-git/go-git/v5"
	. "github.com/onsi/ginkgo/v2"
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
	var tree_replacer = strings.NewReplacer("git@", "https://", ":", "/", ".git", "/tree/")
	var tag_replacer = strings.NewReplacer("git@", "https://", ":", "/", ".git", "/releases/tag/")
	r, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{
                        DetectDotGit: true,
        })
        if err == nil {
		commit_hash := ""
                ref_origin, err_origin := r.Remotes()
		if err_origin == nil {
			origin_url = tree_replacer.Replace(strings.Fields(ref_origin[0].String())[1])
			ref_head, err_head := r.Head()
			if err_head == nil {
				commit_hash = strings.Fields(ref_head.String())[0]
				origin_url = origin_url + commit_hash
			} else {
				GinkgoWriter.Println(err_head)
			}

			tags, tags_err := r.Tags()
			if tags_err == nil {
				for {
					tag, tag_err := tags.Next()
					if tag_err == nil {
						tag_string := tag.Strings()
						if tag_string[1] == commit_hash {
							origin_url = tag_replacer.Replace(strings.Fields(ref_origin[0].String())[1])
							origin_url = origin_url + tag_string[0]
							break
						}
					} else {
						// Expected EOF upon end of Tags iterator
						break
					}
				}
			} else {
				GinkgoWriter.Println("Tags Error:", tags_err)
			}
                } else {
			GinkgoWriter.Println("Remotes Error:", err_origin)
		}
        } else {
		GinkgoWriter.Println("Repository Error:", err)
	}

	if origin_url == "" {
		origin_url = "Not Found"
	}

	return
}
