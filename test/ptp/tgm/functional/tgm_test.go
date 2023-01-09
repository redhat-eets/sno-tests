//go:build tgmfunctionaltests
// +build tgmfunctionaltests

package functional

import (
	"context"
	"flag"
	"os"
	"path"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	"github.com/onsi/ginkgo/v2/types"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/ptp-operator/test/pkg"
	"github.com/redhat-eets/sno-tests/test/pkg/client"
	"github.com/redhat-eets/sno-tests/test/pkg/consts"
	"github.com/redhat-eets/sno-tests/test/pkg/devices"
	"github.com/redhat-eets/sno-tests/test/pkg/k8sreporter"
	"github.com/redhat-eets/sno-tests/test/pkg/namespaces"
	"github.com/redhat-eets/sno-tests/test/pkg/pods"
	_ "github.com/redhat-eets/sno-tests/test/ptp/tgm/functional/tests"
)

var (
	testClient *client.ClientSet
	podOnHost  *corev1.Pod

	OperatorNameSpace = pkg.PtpLinuxDaemonNamespace
	junitPath         *string
	reportPath        *string
	r                 *k8sreporter.KubernetesReporter
)

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
		junitFile := path.Join(*junitPath, "tgm_functional_junit.xml")
		reporterConfig.JUnitReport = junitFile
	}

	if *reportPath != "" {
		kubeconfig := os.Getenv("KUBECONFIG")
		r = k8sreporter.New(kubeconfig, *reportPath, OperatorNameSpace)
	}

	RunSpecs(t, "PTP T-GM Functional Suite", reporterConfig)
}

var _ = BeforeSuite(func() {
	testClient = client.New("")

	// Creating a test namespace for the privileged pod on host that needs
	// different security context than the ptp operator pods and namespace have.
	err := namespaces.Create(consts.TestNamespace, testClient)
	Expect(err).NotTo(HaveOccurred())

	podDef, err := pods.DefinePodOnHost(consts.TestNamespace)
	Expect(err).NotTo(HaveOccurred())

	podOnHost, err = testClient.Pods(consts.TestNamespace).Create(context.Background(), podDef, metav1.CreateOptions{})
	Expect(err).NotTo(HaveOccurred())

	err = pods.WaitForPhase(testClient, podOnHost, corev1.PodRunning, pkg.TimeoutIn3Minutes)
	Expect(err).NotTo(HaveOccurred())

	err = devices.InitWPCDevicePort(testClient, podOnHost)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err := testClient.Pods(consts.TestNamespace).Delete(context.Background(), podOnHost.Name, metav1.DeleteOptions{})
	Expect(err).NotTo(HaveOccurred())

	err = namespaces.Delete(consts.TestNamespace, testClient)
	Expect(err).NotTo(HaveOccurred())
})

var _ = ReportAfterEach(func(specReport types.SpecReport) {
	if specReport.Failed() == false {
		return
	}

	if *reportPath != "" {
		// Fetch pods logs from the past 10 Minutes to be able also get logs from
		// an optional configuration stage that is done before running the suite.
		r.Dump(10*time.Minute, specReport.FullText())
	}
})
