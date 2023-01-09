//go:build multisnotgmtests
// +build multisnotgmtests

package multisno

import (
	"flag"
	"os"
	"path"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"

	"github.com/openshift/ptp-operator/test/pkg"
	"github.com/redhat-eets/sno-tests/test/pkg/k8sreporter"
	_ "github.com/redhat-eets/sno-tests/test/ptp/tgm/multisno/tests"
)

var OperatorNameSpace = pkg.PtpLinuxDaemonNamespace

var junitPath *string
var reportPath *string
var r *k8sreporter.KubernetesReporter

func init() {
	if ns := os.Getenv("OO_INSTALL_NAMESPACE"); len(ns) != 0 {
		OperatorNameSpace = ns
	}

	junitPath = flag.String("junit", "", "the path for the junit format report")
	reportPath = flag.String("report", "", "the path of the report file containing details for failed tests")
}

func TestTGM(t *testing.T) {
	RegisterFailHandler(Fail)

	_, reporterConfig := GinkgoConfiguration()

	if *junitPath != "" {
		junitFile := path.Join(*junitPath, "multi_sno_tgm_junit.xml")
		reporterConfig.JUnitReport = junitFile
	}

	if *reportPath != "" {
		kubeconfig := os.Getenv("KUBECONFIG")
		r = k8sreporter.New(kubeconfig, *reportPath, OperatorNameSpace)
	}

	RunSpecs(t, "Multo SNO PTP T-GM Suite", reporterConfig)
}

var _ = ReportAfterEach(func(specReport types.SpecReport) {
	if specReport.Failed() == false {
		return
	}

	if *reportPath != "" {
		r.Dump(10*time.Minute, specReport.FullText())
	}
})
