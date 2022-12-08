package ptp_test

import (
	"os"
	"strings"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/go-git/go-git/v5"
	client "github.com/redhat-eets/sno-tests/test/pkg/client"
	"gopkg.in/yaml.v3"
)

const (
	defaultCfgDir = "/testconfig"
	cfgFile       = "topology.yaml"
)

var (
	cfgdir    string
	topo      Topology
	clients   map[string]*client.ClientSet
	nodenames map[string]string
	roles     = [...]string{"GM", "Tester"}
)

type Topology struct {
	PTP *struct {
		GM *struct {
			Node       string  `yaml:"node"`
			PortTester *string `yaml:"toTester,omitempty"`
		} `yaml:"gm,omitempty"`
		Tester *struct {
			Node   string `yaml:"node"`
			PortGM string `yaml:"toGM"`
		} `yaml:"tester,omitempty"`
	} `yaml:"ptp,omitempty"`
}

func TestPtp(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ptp Suite")
}

var replacer = strings.NewReplacer("git@", "https://", ":", "/", ".git", "/tree/")

var _ = BeforeSuite(func() {
	// Get the current git hash and link
	r, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err == nil {
		ref_head, err_head := r.Head()
		ref_origin, err_origin := r.Remotes()
		if err_head == nil && err_origin == nil {
			commit_hash := strings.Fields(ref_head.String())[0]
			origin_url := replacer.Replace(strings.Fields(ref_origin[0].String())[1])
			origin_url = origin_url + commit_hash
			GinkgoWriter.Println("Local Test Suite Link: ", origin_url)
		}
	}
	// Get the config file location from enviroment
	val, ok := os.LookupEnv("CFG_DIR")
	if !ok {
		cfgdir = defaultCfgDir
	} else {
		cfgdir = val
	}

	cfg := cfgdir + "/" + cfgFile
	yfile, err := os.ReadFile(cfg)
	Expect(err).NotTo(HaveOccurred())

	Expect(yaml.Unmarshal(yfile, &topo)).To(Succeed())

	// Skip this test suite if no ptp topology is specified
	if topo.PTP == nil {
		Skip("PTP test suite not requested.")
	}

	// Setup k8s api client for each cluster under PTP section
	clients = make(map[string]*client.ClientSet)
	// clusters can be shared by multiple nodes, where clients is per node
	clusters := make(map[string]*client.ClientSet)
	nodenames = make(map[string]string)
	info := make(map[string][]string)
	if topo.PTP.GM != nil && topo.PTP.GM.Node != "" {
		info["GM"] = strings.Split(topo.PTP.GM.Node, "/")
	}

	if topo.PTP.Tester != nil && topo.PTP.Tester.Node != "" {
		info["Tester"] = strings.Split(topo.PTP.Tester.Node, "/")
	}

	for _, role := range roles {
		if _, ok := info[role]; !ok {
			clients[role] = nil
			nodenames[role] = ""
			continue
		}

		if len(info[role]) > 1 {
			nodenames[role] = info[role][1]
		} else {
			nodenames[role] = ""
		}
		kubecfg := cfgdir + "/" + info[role][0]
		if _, ok := clusters[kubecfg]; !ok {
			clusters[kubecfg] = client.New(kubecfg)
		}
		clients[role] = clusters[kubecfg]
	}

})

var _ = AfterEach(func() {

})
