---
title: Virtual Machines
publishDate: "2019-12-31"
categories: ["Traffic Management"]
---

Running containerized services in Kubernetes can add lots of benefits, including autoscaling, dependency isolation, and resource optimization. Adding Istio to your Kubernetes environment can radically simplify metrics aggregation and policy management, even if you're operating lots of containers.

But what about stateful services, or legacy applications are running in virtual machines? Or what if you're migrating from VMs to containers? Have no fear: you can integrate virtual machines into your Istio service mesh. Let's see how.

![](/images/vm-call-flow.png)

In this example, we are running a web application for a local library. This web app has multiple backends, all running in a Kubernetes cluster. One of our Kubernetes workloads, `inventory`, talks to a PostgreSQL database, and writes a record for each new book added to the library. This database is running in a virtual machine in another cloud region.

In order to get the full Istio functionality for PostgreSQL, our VM-based database, we need to install the Istio sidecar proxy on the VM, and configure it to talk to the Istio control plane running in the cluster. (Note that this is different from adding external [ServiceEntries](/external-services).) We can add our Postgres database to the mesh with three steps. You can follow along with the commands in the [demo on GitHub](https://github.com/askmeegs/postgres-library/tree/0241acce9d7e2cede0de8ac9baa1338624f716eb#-postgres-library).

![](/images/vm-architecture.png)

1. **[Create a firewall rule](https://github.com/askmeegs/postgres-library#5-allow-pod-ip-traffic-to-the-vm) for pod-to-VM traffic**. This allows traffic from the Kubernetes pod CIDR range directly into our VM-based workload.

2. **[Install Istio](https://github.com/askmeegs/postgres-library#6-prepare-a-clusterenv-file-to-send-to-the-vm) on the VM**. Copy the service account keys, and the ports the VM service exposes (in this case, PostgreSQL client port `5432`). Update the `/etc/hosts` on the VM to route `istio.pilot` and `istio.citadel` traffic through the Istio IngressGateway running on the cluster. Then, install the Istio sidecar proxy and node agent on the VM. The node agent is responsible for generating client certificates to mount into the sidecar proxy, for mutual TLS authentication. Start the proxy and node agent using `systemctl`.

3. **[Register the VM](https://github.com/askmeegs/postgres-library#8-register-the-vm-with-istio-running-on-the-gke-cluster) workload with Kubernetes**. We need two Kubernetes resources to add our Postgres database to the mesh. The first is a `ServiceEntry`. This will route the Kubernetes DNS name to the virtual machine IP address. Finally, we need a Kubernetes `Service` to create that DNS entry. This is what will allow our client pod to [start a database connection](https://github.com/askmeegs/postgres-library/blob/master/main.go#L39) at `postgres-1-vm.default.svc.cluster.local`. We can use the `istioctl register` command to do this.

We can now verify that the Kubernetes-based client can successfully write to the database, by viewing the pod logs:

```
postgres-library-6bb956f86b-dt94x server ✅ inserted Fun Home
postgres-library-6bb956f86b-dt94x server ✅ inserted Infinite Jest
postgres-library-6bb956f86b-dt94x server ✅ inserted To the Lighthouse
```

And we can verify that the sidecar proxy running on the VM is intercepting inbound traffic on port `5432`, by viewing the Envoy access logs on the VM.

```
$ tail /var/log/istio/istio.log

[2019-11-14T19:09:00.174Z] "- - -" 0 - "-" "-" 268 441 194 - "-" "-" "-" "-"
"127.0.0.1:5432" inbound|5432|tcp|postgresql-1-vm.default.svc.cluster.local
127.0.0.1:54104 10.128.0.14:5432 10.24.2.23:40190
outbound_.5432_._.postgresql-1-vm.default.svc.cluster.local -
```

We can also see TCP metrics flowing in the Kiali service graph:

![](/images/vm-kiali.png)


From here, we can use any Istio traffic or security policies to our multi-environment service mesh. For instance, we can encrypt all traffic between the cluster and the VM, by adding a mesh-wide mutual TLS policy:

```YAML
apiVersion: "authentication.istio.io/v1alpha1"
kind: "MeshPolicy"
metadata:
  name: "default"
spec:
  peers:
  - mtls: {}
---
apiVersion: "networking.istio.io/v1alpha3"
kind: "DestinationRule"
metadata:
  name: "default"
  namespace: "istio-system"
spec:
  host: "*.local"
  trafficPolicy:
    tls:
      mode: ISTIO_MUTUAL
```


To learn more:
- [Run this example in your environment](https://github.com/askmeegs/postgres-library)
- [Try another example](https://github.com/GoogleCloudPlatform/istio-samples/tree/master/mesh-expansion-gce)
- [See the Istio documentation](https://istio.io/docs/examples/virtual-machines/single-network/)