# sno-tests

## PTP T-GM Suite

### Prerequisites

- You need at least one running SNO with a WPC NIC and ICE driver installed.
- You need to install the ptp-operator and set a grandmaster PtpConfig.

### PTP T-GM Functionality Tests

#### Running tests directly from git source tree

First, git clone the source tree.

Install Ginkgo and Gomega:
```
go install github.com/onsi/ginkgo/v2/ginkgo
go get github.com/onsi/gomega/...
```

- Set the `KUBECONFIG` variable.
- Optionally set the `TESTS_REPORTS_PATH` variable.

Run T-GM functionality tests:
```
make test-tgm
```

Optionally run validation tests only:
```
make test-tgm-validation-only
```

### PTP T-GM Multi SNO Tests

#### DNS entries on the test machine

In order for the script to control multiple SNO clusters, the test machine where the script is running should be able to resolve the DNS names for the different SNO clusters. For example, here are sample `/etc/hosts` entries on the test machine, added in addition to the pre-existing entries,	

```	
$ cat /etc/hosts	
192.168.49.151	api.node2.wpc.test	
192.168.49.151	oauth-openshift.apps.node2.wpc.test	
192.168.49.151	console-openshift-console.apps.node2.wpc.test	
192.168.49.151	grafana-openshift-monitoring.apps.node2.wpc.test	
192.168.49.151	thanos-querier-openshift-monitoring.apps.node2.wpc.test	
192.168.49.151	prometheus-k8s-openshift-monitoring.apps.node2.wpc.test	
192.168.49.151	alertmanager-main-openshift-monitoring.apps.node2.wpc.test	
192.168.49.141	api.node12.wpc.test	
192.168.49.141	oauth-openshift.apps.node12.wpc.test	
192.168.49.141	console-openshift-console.apps.node12.wpc.test	
192.168.49.141	grafana-openshift-monitoring.apps.node12.wpc.test	
192.168.49.141	thanos-querier-openshift-monitoring.apps.node12.wpc.test	
192.168.49.141	prometheus-k8s-openshift-monitoring.apps.node12.wpc.test	
192.168.49.141	alertmanager-main-openshift-monitoring.apps.node12.wpc.test	
```	

Run T-GM multi SNO tests:
```
make test-tgm-multisno
```

### Running tests from a test container	

The Ginkgo test suites can be built into a binary and run from inside a container.	

Here is an example to build the test container image and push to a private docker image repository, from the git tree root directory, 	
```	
podman build -t 192.168.49.147:5000/ptp-test .	
podman push 192.168.49.147:5000/ptp-test --tls-verify=false	
```	

In order to be able to resolve the DNS query for multiple SNO clusters when the test container is running from an OpenShift cluster, extra DNS entries can be added to the `/etc/hosts` inside the container. Basically, the test container needs to be able access the following information:	
* kubeconfig file
* extra DNS entries for each SNO cluster and saved in `dns-entries`	

A config map can be used to pass in all this information, as in the below example,	
```	
# ls ptp-testconfig	
dns-entries  kubeconfig
# oc create configmap test-config --from-file=ptp-testconfig	
```	

In the above steps, we can see there is a extra file called `dns-entries` under the directory that's used to build the config map. Let's take look of its content,	
```	
# cat ptp-testconfig/dns-entries	
192.168.49.151	api.node2.wpc.test	
192.168.49.151	oauth-openshift.apps.node2.wpc.test	
192.168.49.151	console-openshift-console.apps.node2.wpc.test	
192.168.49.151	grafana-openshift-monitoring.apps.node2.wpc.test	
192.168.49.151	thanos-querier-openshift-monitoring.apps.node2.wpc.test	
192.168.49.151	prometheus-k8s-openshift-monitoring.apps.node2.wpc.test	
192.168.49.151	alertmanager-main-openshift-monitoring.apps.node2.wpc.test	
192.168.49.141	api.node12.wpc.test	
192.168.49.141	oauth-openshift.apps.node12.wpc.test	
192.168.49.141	console-openshift-console.apps.node12.wpc.test	
192.168.49.141	grafana-openshift-monitoring.apps.node12.wpc.test	
192.168.49.141	thanos-querier-openshift-monitoring.apps.node12.wpc.test	
192.168.49.141	prometheus-k8s-openshift-monitoring.apps.node12.wpc.test	
192.168.49.141	alertmanager-main-openshift-monitoring.apps.node12.wpc.test	
```	

Essentially these entries are the same as the extra `/etc/hosts` entries necessary when we run the "ginkgo" tests directly from git source tree.	

This is a sample YAML file to run the test container image inside openshift,	
```	
# the configmap test-config is created with	
# oc create configmap test-config --from-file=<dir>	
apiVersion: v1 	
kind: Pod 	
metadata:	
  name: sno-tests	
spec:	
  restartPolicy: Never	
  containers:	
  - name: sno-tests 	
    image: 192.168.49.147:5000/ptp-test:latest	
    securityContext:	
      privileged: true	
    volumeMounts:	
      - name: config-volume	
        mountPath: /testconfig	
  volumes:	
    - name: config-volume	
      configMap:	
        name: test-config	
```	

After the pods run and complete, get the test result via the pod log,	
```	
# oc get pods	
NAME        READY   STATUS      RESTARTS   AGE	
sno-tests   0/1     Completed   0          31h	
# oc logs sno-tests	
=== RUN   TestPtp	
Running Suite: Ptp Suite - /	
============================	
Random Seed: 1670418609	
Will run 3 of 3 specs	
time="2022-12-07T13:10:09Z" level=info msg="number of ptpPods: 1"	
time="2022-12-07T13:10:09Z" level=info msg="number of ptpRunningPods: 1"	
time="2022-12-07T13:10:11Z" level=info msg="captured log: $GNRMC,131010.00,A,4233.01508,N,07112.87791,W,0.009,,071222,,,A,V*06\r\r\n$GNGGA,131010.00,4233.01508,N,07112.87791,W,1,07,1.18,57.3,M,-33.0,M,,*4E\r\r\n$GNGGA,131011.00,4233.01509,N,07112.87791,W,1,07,1.18,57.3,M,-33.0,M,,*4E\r\r\n"	
•time="2022-12-07T13:10:11Z" level=info msg="number of ptpPods: 1"	
time="2022-12-07T13:10:11Z" level=info msg="number of ptpRunningPods: 1"	
time="2022-12-07T13:10:15Z" level=info msg="captured log: ,131015.00,A,4233.01509,N,07112.87788,W,0.009,,071222,,,A,V"	
•time="2022-12-07T13:10:15Z" level=info msg="number of ptpPods: 1"	
time="2022-12-07T13:10:15Z" level=info msg="number of ptpRunningPods: 1"	
time="2022-12-07T13:10:20Z" level=info msg="captured log: 388531 max 1228470 freq +5777854 +/- 5009763 delay   947 +/- 236"	
•	
Ran 3 of 3 Specs in 10.703 seconds	
SUCCESS! -- 3 Passed | 0 Failed | 0 Pending | 0 Skipped	
--- PASS: TestPtp (10.70s)	
PASS	
```
