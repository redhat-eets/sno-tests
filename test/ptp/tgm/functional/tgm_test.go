//go:build tgmfunctionaltests
// +build tgmfunctionaltests

package functional

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/openshift/ptp-operator/test/pkg"
	"github.com/redhat-eets/sno-tests/test/pkg/client"
	"github.com/redhat-eets/sno-tests/test/pkg/consts"
	"github.com/redhat-eets/sno-tests/test/pkg/devices"
	"github.com/redhat-eets/sno-tests/test/pkg/namespaces"
	"github.com/redhat-eets/sno-tests/test/pkg/pods"
	_ "github.com/redhat-eets/sno-tests/test/ptp/tgm/functional/tests"
)

var (
	testClient *client.ClientSet
	podOnHost  *corev1.Pod
)

func TestTGM(t *testing.T) {
	RegisterFailHandler(Fail)
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
