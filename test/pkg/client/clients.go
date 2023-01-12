package client

import (
	"os"

	"github.com/golang/glog"

	clientconfigv1 "github.com/openshift/client-go/config/clientset/versioned/typed/config/v1"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	appsv1client "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	networkv1client "k8s.io/client-go/kubernetes/typed/networking/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ptpapiv1 "github.com/openshift/ptp-operator/api/v1"
	ptpv1 "github.com/openshift/ptp-operator/pkg/client/clientset/versioned/typed/ptp/v1"
)

type ClientSet struct {
	client.Client
	kubernetes.Interface
	networkv1client.NetworkingV1Client
	appsv1client.AppsV1Interface
	ptpv1.PtpV1Interface
	Config    *rest.Config
	OcpClient clientconfigv1.ConfigV1Interface
	corev1client.CoreV1Interface
}

func New(kubeconfig string) *ClientSet {
	var config *rest.Config
	var err error

	if kubeconfig == "" {
		kubeconfig = os.Getenv("KUBECONFIG")
	}

	if kubeconfig != "" {
		glog.V(4).Infof("Loading kube client config from path %q", kubeconfig)
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		glog.V(4).Infof("Using in-cluster kube client config")
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		glog.Infof("Failed to create a valid client,  environement variable.")
		// Cannot create client, nothing else to do
		os.Exit(1)
	}

	clientSet := &ClientSet{}
	clientSet.CoreV1Interface = corev1client.NewForConfigOrDie(config)
	clientSet.Interface = kubernetes.NewForConfigOrDie(config)
	clientSet.AppsV1Interface = appsv1client.NewForConfigOrDie(config)
	clientSet.NetworkingV1Client = *networkv1client.NewForConfigOrDie(config)
	clientSet.PtpV1Interface = ptpv1.NewForConfigOrDie(config)
	clientSet.OcpClient = clientconfigv1.NewForConfigOrDie(config)
	clientSet.Config = config

	myScheme := runtime.NewScheme()
	if err = scheme.AddToScheme(myScheme); err != nil {
		panic(err)
	}

	if err := apiext.AddToScheme(myScheme); err != nil {
		panic(err)
	}

	if err := ptpapiv1.AddToScheme(myScheme); err != nil {
		panic(err)
	}

	clientSet.Client, err = client.New(config, client.Options{
		Scheme: myScheme,
	})

	if err != nil {
		return nil
	}

	return clientSet
}
