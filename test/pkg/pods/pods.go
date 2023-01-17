package pods

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"github.com/onsi/gomega"
	"github.com/openshift/ptp-operator/test/pkg"
	"github.com/openshift/ptp-operator/test/pkg/images"
	"github.com/redhat-eets/sno-tests/test/pkg/client"
	"github.com/sirupsen/logrus"
	"github.com/test-network-function/l2discovery-lib/pkg/pods"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/utils/pointer"
)

// GetLog connects to a pod and fetches log
func GetLog(api *client.ClientSet, p *corev1.Pod, containerName string) (string, error) {
	req := api.CoreV1().Pods(p.Namespace).GetLogs(p.Name, &corev1.PodLogOptions{Container: containerName})
	log, err := req.Stream(context.Background())
	if err != nil {
		return "", err
	}
	defer log.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, log)

	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// ExecCommand runs command in the pod and returns buffer output
func ExecCommand(api *client.ClientSet, pod *corev1.Pod, containerName string, command []string) (bytes.Buffer, error) {
	var buf bytes.Buffer
	req := api.CoreV1().RESTClient().
		Post().
		Namespace(pod.Namespace).
		Resource("pods").
		Name(pod.Name).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(api.Config, "POST", req.URL())
	if err != nil {
		return buf, err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: &buf,
		Stderr: os.Stderr,
		Tty:    true,
	})

	return buf, err
}

// returns true if the pod passed as paremeter is running on the node selected by the label passed as a parameter.
// the label represent a ptp conformance test role such as: grandmaster, clock under test, slave1, slave2
func PodRole(api *client.ClientSet, runningPod *corev1.Pod, label string) (bool, error) {
	nodeList, err := api.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{
		LabelSelector: label,
	})
	if err != nil {
		return false, fmt.Errorf("error getting node list")
	}
	for NodeNumber := range nodeList.Items {
		if runningPod.Spec.NodeName == nodeList.Items[NodeNumber].Name {
			return true, nil
		}
	}
	return false, nil
}

// returns true if a pod has a given label or node name
func HasPodLabelOrNodeName(api *client.ClientSet, pod *corev1.Pod, label *string, nodeName *string) (result bool, err error) {
	if label == nil && nodeName == nil {
		return result, fmt.Errorf("label and nodeName are nil")
	}
	// node name might be present and will be superseded by label
	/*if label != nil && nodeName != nil {
		return result, fmt.Errorf("label or nodeName must be nil")
	}*/
	if label != nil {
		result, err = PodRole(api, pod, *label)
		if err != nil {
			return result, fmt.Errorf("could not check %s pod role, err: %s", *label, err)
		}
	}
	if nodeName != nil {
		result = pod.Spec.NodeName == *nodeName
	}
	return result, nil
}

// WaitForCondition waits until the pod will have specified condition type with the expected status
func WaitForCondition(cs *client.ClientSet, pod *corev1.Pod, conditionType corev1.PodConditionType, conditionStatus corev1.ConditionStatus, timeout time.Duration) error {
	return wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		updatePod, err := cs.Pods(pod.Namespace).Get(context.Background(), pod.Name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		for _, c := range updatePod.Status.Conditions {
			if c.Type == conditionType && c.Status == conditionStatus {
				return true, nil
			}
		}
		return false, nil
	})
}

// WaitForPhase waits until the pod will be in specified phase
func WaitForPhase(cs *client.ClientSet, pod *corev1.Pod, phaseType corev1.PodPhase, timeout time.Duration) error {
	return wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		updatePod, err := cs.Pods(pod.Namespace).Get(context.Background(), pod.Name, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}

		return updatePod.Status.Phase == phaseType, nil
	})
}

func WaitUntilLogIsDetected(api *client.ClientSet, pod *corev1.Pod, timeout time.Duration, neededLog string) {
	gomega.Eventually(func() string {
		logs, _ := GetLog(api, pod, pkg.PtpContainerName)
		logrus.Debugf("wait for log = %s in pod=%s.%s", neededLog, pod.Namespace, pod.Name)
		return logs
	}, timeout, 1*time.Second).Should(gomega.ContainSubstring(neededLog), fmt.Sprintf("Timeout to detect log %q in pod %q", neededLog, pod.Name))
}

// looks for a given pattern in a pod's log and returns when found
func WaitUntilLogIsDetectedRegex(api *client.ClientSet, pod *corev1.Pod, timeout time.Duration, regex string) string {
	var results []string
	gomega.Eventually(func() []string {
		podLogs, _ := GetLog(api, pod, pkg.PtpContainerName)
		logrus.Debugf("wait for log = %s in pod=%s.%s", regex, pod.Namespace, pod.Name)
		r := regexp.MustCompile(regex)
		var id string

		for _, submatches := range r.FindAllStringSubmatchIndex(podLogs, -1) {
			id = string(r.ExpandString([]byte{}, "$1", podLogs, submatches))
			results = append(results, id)
		}
		return results
	}, timeout, 5*time.Second).Should(gomega.Not(gomega.HaveLen(0)), fmt.Sprintf("Timeout to detect regex %q in pod %q", regex, pod.Name))
	if len(results) != 0 {
		return results[len(results)-1]
	}
	return ""
}

func CheckRestart(pod corev1.Pod) {
	logrus.Printf("Restarting the node %s that pod %s is running on", pod.Spec.NodeName, pod.Name)

	const (
		pollingInterval = 3 * time.Second
	)

	gomega.Eventually(func() error {
		_, err := pods.ExecCommand(&pod, "container-00", []string{"chroot", "/host", "shutdown", "-r"})
		return err
	}, pkg.TimeoutIn10Minutes, pollingInterval).Should(gomega.BeNil())
}

func GetRebootDaemonsetPodsAt(api *client.ClientSet, node string) *corev1.PodList {

	rebootDaemonsetPodList, err := api.CoreV1().Pods(pkg.RebootDaemonSetNamespace).List(context.Background(), metav1.ListOptions{LabelSelector: "name=" + pkg.RebootDaemonSetName, FieldSelector: fmt.Sprintf("spec.nodeName=%s", node)})
	gomega.Expect(err).ToNot(gomega.HaveOccurred())

	return rebootDaemonsetPodList
}

func getDefinition(namespace string) *corev1.Pod {
	podObject := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "testpod-",
			Namespace:    namespace},
		Spec: corev1.PodSpec{
			TerminationGracePeriodSeconds: pointer.Int64(0),
			Containers: []corev1.Container{{Name: "test",
				Image:   images.For(images.TestUtils),
				Command: []string{"/bin/bash", "-c", "sleep INF"}}}}}

	return podObject
}

// DefinePodOnNode creates the pod defintion with a node selector
func DefinePodOnNode(namespace string, nodeName string) *corev1.Pod {
	pod := getDefinition(namespace)
	pod.Spec.NodeSelector = map[string]string{"kubernetes.io/hostname": nodeName}
	return pod
}

// DefinePodOnHost creates the pod defintion with a mount to host
func DefinePodOnHost(namespace string) (*corev1.Pod, error) {
	pod := getDefinition(namespace)

	pod = RedefineWithMount(pod, pod.Spec.Containers[0].Name,
		corev1.Volume{
			Name: "host",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/",
				},
			}},
		corev1.VolumeMount{
			Name:      "host",
			MountPath: "/host"})

	pod, err := RedefineAsPrivileged(pod, pod.Spec.Containers[0].Name)
	if err != nil {
		return nil, err
	}

	return pod, nil
}

// RedefineWithMount updates the container definition with a volume and volume mount
func RedefineWithMount(pod *corev1.Pod, containerName string, volume corev1.Volume, mount corev1.VolumeMount) *corev1.Pod {
	for i, c := range pod.Spec.Containers {
		if c.Name == containerName {
			pod.Spec.Containers[i].VolumeMounts = []corev1.VolumeMount{mount}
			pod.Spec.Volumes = []corev1.Volume{volume}
		}
	}

	return pod
}

// RedefineAsPrivileged updates the container definition to be privileged
func RedefineAsPrivileged(pod *corev1.Pod, containerName string) (*corev1.Pod, error) {
	b := true
	for i, c := range pod.Spec.Containers {
		if c.Name == containerName {
			pod.Spec.Containers[i].SecurityContext = &corev1.SecurityContext{
				Privileged: &b,
			}
		}
	}

	return pod, nil
}

// GetPTPDaemonPod returns the ptp daemon pod running on the SNO.
// returns the first element from the pods list since there is only pod, one daemon per node.
func GetPTPDaemonPod(api *client.ClientSet) (*corev1.Pod, error) {
	ptpRunningPods := []*corev1.Pod{}

	ptpPods, err := api.CoreV1().Pods(pkg.PtpLinuxDaemonNamespace).List(context.Background(), metav1.ListOptions{LabelSelector: "app=linuxptp-daemon"})
	if err != nil {
		return nil, err
	}

	if len(ptpPods.Items) == 0 {
		return nil, fmt.Errorf("linuxptp-daemon is not deployed on cluster")
	}

	for podIndex := range ptpPods.Items {
		ptpRunningPods = append(ptpRunningPods, &ptpPods.Items[podIndex])
	}

	return ptpRunningPods[0], nil
}
