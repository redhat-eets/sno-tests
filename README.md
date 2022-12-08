# sno-tests

## DNS entries on the test machine

In order for the script to control multple SNO clusters, the test machine where the script is running should be able to resolve the DNS names for the different SNO clusters, for example, here is a sample `/etc/hosts` extra entries on the test machine, in addition to the pre-existing entries,

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

## Running test directly from git source tree

First git clone the source tree.

Install  ginkgo and gomega,

```
go install github.com/onsi/ginkgo/v2/ginkgo
go get github.com/onsi/gomega/...
```

Create a directory to host the SNO kubeconfig topology files. By default the ginkgo script expects to find these files under `/testconfig` directory. This directory path can be changed by the enviroment var `CFG_DIR`.

Here is an example list of files under this configuration directory,
```
# ls /testconfig/
node12  node2  topology.yaml
```

In the above listing, the `node12` file and `node2` file are the kubeconfig files for SNO node12 and SNO node2. These kubeconfig file name can be different than the node name, for example, `cluster12` for node12, `cluster2` for node2.  The `topology.yaml` is the file is used to describe the roles and port connections, and this file name can not be changed.

Let's take a look of the content of a sample `topology.yaml`:
```
ptp:
  gm:
    node: node12
    toTester: ens7f3
  tester:
    node: node2
    toGM: ens2f3
```

This describe two roles will be used for the PTP tests, "gm" role and "tester" role. The "gm" role should be straghtforward - a T-GM. The "tester" role needs some explanation - just like the "gm", the "tester" also has a WPC NIC installed and with GNSS connected. The "tester" can be running in "free run" mode and uses its GPS to check the T-GM PTP timing accuracy. The "tester" can also be running the same functionality as a normal "slave" clock and sync up to the "gm" via PTP.

Each role has some ports connecting to other roles. Under the "gm", `toTester` means which port on the "gm" is used to connect the "tester"; under the "tester", `toGM` means which port on the "tester" is used to connect the "gm".

If the test enviroment does not have a second server with the GNSS connection, that's fine - certain tests cases will simply be skipped.

If the test enviroment has only one single SNO and still want to test the WPC GNSS function on that single SNO, it can be done with a sample `topology.yaml` like below,
```
ptp:
  gm:
    node: node12
    toTester: ens7f3    # A WPC port is needed even if the ethernet port is not connected
```

In the above topology file, there is one single gm role for the single SNO. In order to test the GNSS, `toTester` is still specified, even though no "tester" is defined. The `toTester` ethernet port does not need to be connected or up. The ginkgo script uses the `toTester` to find the intended WPC PCI address and query the GNSS module.  

With the above files under the `/testconfig` directory, the ginkgo test suite can be triggered,
```
 ginkgo -v
 ```

## Running test from a test container

The ginkgo test suites can be built into a binary and run from inside a container.

Here is an example to build the test container image and push to a private docker image repository, from the git tree root directory, 
```
podman build -t 192.168.49.147:5000/ptp-test .
podman push 192.168.49.147:5000/ptp-test --tls-verify=false
```

When the test container is running from an OpenShift cluster, in order to be able to resolve the DNS query for multiple SNO clusters, extra DNS entries can be added to the `/etc/hosts` inside the container. Basically the test container need to be able access the following information:
* kubeconfig file for each SNO custer
* toplogy file
* extra DNS entries for each SNO cluster and saved in `dns-entries`

A config map can be used to pass in all these information.

Here is an example to build this config map.
```
# ls ptp-testconfig
dns-entries  node12  node2  topology.yaml

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

Essentially these entries are the same as the extra /etc/hosts entries when we run the "ginkgo" test directly from git source tree.

This is an sample yaml file to run the test container image inside openshift,
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

After the pods run and completes, get the test result via the pod log,
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