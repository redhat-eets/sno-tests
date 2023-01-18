// SPDX-License-Identifier:Apache-2.0

package k8sreporter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/onsi/ginkgo/v2"
	"github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	ptpv1 "github.com/openshift/ptp-operator/api/v1"
	"github.com/redhat-eets/sno-tests/test/pkg/client"
)

const fileSeparator = "-----------------------------------\n"

// KubernetesReporter is a Ginkgo reporter that dumps info
// about configured kubernetes objects.
type KubernetesReporter struct {
	sync.Mutex
	clients        *client.ClientSet
	reportPath     string
	namespaceToLog string
	crs            []CRData
}

// CRData represents a cr to dump
type CRData struct {
	Cr        runtimeclient.ObjectList
	Namespace *string
}

// New returns a new Kubernetes reporter from the given configuration.
func New(kubeconfig, path, namespace string) *KubernetesReporter {
	// The list of CRDs we want to dump
	crds := []CRData{
		{Cr: &ptpv1.PtpConfigList{}},
		{Cr: &ptpv1.PtpOperatorConfigList{}},
		{Cr: &ptpv1.NodePtpDeviceList{}},
		{Cr: &corev1.ServiceList{}},
	}

	return &KubernetesReporter{
		clients:        client.New(kubeconfig),
		reportPath:     path,
		namespaceToLog: namespace,
		crs:            crds,
	}
}

// Dump dumps the relevant crs + pod logs.
// duration represents how much in the past we need to go when fetching the pod
// logs.
// dumpSubpath is the subpath relative to reportPath where the reporter will
// dump the output.
func (r *KubernetesReporter) Dump(duration time.Duration, testName string) {
	dumpSubpath := strings.Replace(ginkgo.CurrentGinkgoTestDescription().TestText, " ", "-", -1)

	since := time.Now().Add(-duration).Add(-5 * time.Second)

	err := os.Mkdir(path.Join(r.reportPath, dumpSubpath), 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		logrus.Infof("failed to create test dir: %v\n", err)
		return
	}
	r.logNodes(dumpSubpath)
	r.logLogs(since, dumpSubpath)
	r.logPods(dumpSubpath)

	for _, cr := range r.crs {
		r.logCustomCR(cr.Cr, cr.Namespace, dumpSubpath)
	}
}

// logPods dumping pods info into a file in the reporting path.
func (r *KubernetesReporter) logPods(dirName string) {
	pods, err := r.clients.Pods(corev1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logrus.Infof("failed to fetch pods: %v\n", err)
		return
	}
	for _, pod := range pods.Items {
		if r.namespaceToLog != pod.Namespace {
			continue
		}
		f, err := logFileFor(r.reportPath, dirName, pod.Namespace+"-pods_specs")
		if err != nil {
			logrus.Infof("failed to open pods_specs file: %v\n", dirName)
			return
		}
		defer f.Close()
		fmt.Fprintf(f, fileSeparator)
		j, err := json.MarshalIndent(pod, "", "    ")
		if err != nil {
			fmt.Println("Failed to marshal pods", err)
			return
		}
		fmt.Fprintln(f, string(j))
	}
}

// logNodes dumping nodes info into a file in the reporting path.
func (r *KubernetesReporter) logNodes(dirName string) {
	f, err := logFileFor(r.reportPath, dirName, "nodes")
	if err != nil {
		logrus.Infof("failed to open nodes file: %v\n", dirName)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, fileSeparator)

	nodes, err := r.clients.Nodes().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logrus.Infof("failed to fetch nodes: %v\n", err)
		return
	}

	j, err := json.MarshalIndent(nodes, "", "    ")
	if err != nil {
		fmt.Println("Failed to marshal nodes")
		return
	}
	fmt.Fprintln(f, string(j))
}

// logLogs fetching pods logs into a file in the reporting path.
func (r *KubernetesReporter) logLogs(since time.Time, dirName string) {
	pods, err := r.clients.Pods(corev1.NamespaceAll).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		logrus.Infof("failed to fetch pods: %v\n", err)
		return
	}
	for _, pod := range pods.Items {
		if r.namespaceToLog != pod.Namespace {
			continue
		}
		f, err := logFileFor(r.reportPath, dirName, pod.Namespace+"-pods_logs")
		if err != nil {
			logrus.Infof("failed to open pods_logs file: %v\n", dirName)
			return
		}
		defer f.Close()
		containersToLog := make([]corev1.Container, 0)
		containersToLog = append(containersToLog, pod.Spec.Containers...)
		containersToLog = append(containersToLog, pod.Spec.InitContainers...)
		for _, container := range containersToLog {
			logStart := metav1.NewTime(since)
			logs, err := r.clients.Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{Container: container.Name, SinceTime: &logStart}).DoRaw(context.Background())
			if err == nil {
				fmt.Fprintf(f, fileSeparator)
				fmt.Fprintf(f, "Dumping logs for pod %s-%s-%s\n", pod.Namespace, pod.Name, container.Name)
				fmt.Fprintln(f, string(logs))
			}
		}

	}
}

// logCustomCR dumping custom CRs info into a file in the reporting path.
func (r *KubernetesReporter) logCustomCR(cr runtimeclient.ObjectList, namespace *string, dirName string) {
	f, err := logFileFor(r.reportPath, dirName, "crs")
	if err != nil {
		logrus.Infof("failed to open crs file: %v\n", dirName)
		return
	}
	defer f.Close()
	fmt.Fprintf(f, fileSeparator)
	if namespace != nil {
		fmt.Fprintf(f, "Dumping %T in namespace %s\n", cr, *namespace)
	} else {
		fmt.Fprintf(f, "Dumping %T\n", cr)
	}

	options := []runtimeclient.ListOption{}
	if namespace != nil {
		options = append(options, runtimeclient.InNamespace(*namespace))
	}
	err = r.clients.List(context.Background(),
		cr,
		options...)

	if err != nil {
		// this can be expected if we are reporting a feature we did not install the operator for
		fmt.Fprintf(f, "Failed to fetch %T: %v\n", cr, err)
		return
	}

	j, err := json.MarshalIndent(cr, "", "    ")
	if err != nil {
		fmt.Fprintf(f, "Failed to marshal %T\n", cr)
		return
	}
	fmt.Fprintln(f, string(j))
}

// logFileFor creates a file in the reporting path.
func logFileFor(dirName string, testName string, kind string) (*os.File, error) {
	path := path.Join(dirName, testName, kind) + ".log"
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return f, nil
}
