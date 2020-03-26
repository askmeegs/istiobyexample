---
title: Database Traffic
lastmod: "2019-12-31"
publishDate: "2019-12-31"
categories: ["Traffic Management"]
---


Applications often span multiple environments, and databases are a great example. You might choose to run your database [outside of Kubernetes](https://cloud.google.com/blog/products/databases/to-run-or-not-to-run-a-database-on-kubernetes-what-to-consider) for legacy or storage reasons, or you might use a managed database service.

But fear not! You can still add external databases to your Istio service mesh. Let's see how.

![diagram](/images/databases-diagram.png)

Here, we have a `plants` service running inside a Kubernetes cluster, with Istio enabled. `plants` writes inventory to a [Firestore](https://firebase.google.com/docs/firestore) NoSQL database running in Google Cloud, using the Golang client library for Firestore. Its logs look like this:

```bash
writing a new plant to Firestore...
✅success
```

Let's say we want to monitor outgoing traffic to Firestore. To do this, we'll add an Istio [ServiceEntry](https://istio.io/docs/reference/config/networking/v1alpha3/service-entry/) corresponding to the hostname of the [Firestore API](https://cloud.google.com/firestore/docs/reference/rpc/).

```YAML
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: firestore
spec:
  hosts:
  - "firestore.googleapis.com"
  ports:
  - name: https
    number: 443
    protocol: HTTPS
  location: MESH_EXTERNAL
  resolution: DNS
```

From here, we can see Firestore appear in Istio's [service graph](https://istio.io/docs/tasks/telemetry/kiali/).

![kiali](/images/databases-kiali-no-vs.png)

Note that the traffic appears as TCP because the sidecar proxy for `plants` is receiving the firestore TLS traffic [as plain TCP](https://github.com/istio/istio/issues/14933). The edge of the graph denote the number of request throughput to Firestore, in bits per second.

Now let's say we want to test `plants`'s behavior when it cannot connect to the database. We can do this with Istio, without changing any application code.

While Istio currently does not support TCP fault injection, what we can do is create a [TCP traffic rule](https://istio.io/docs/reference/config/networking/v1alpha3/virtual-service/#TCPRoute) to send firestore API traffic to another "black hole" service, effectively breaking the client connection to Firestore.

To do this, we can deploy a small `echo` service inside the cluster, and forward all `firestore` traffic to the `echo` service instead:

```YAML
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: fs
spec:
  hosts:
  - firestore.googleapis.com
  tcp:
  - route:
    - destination:
        host: echo
        port:
          number: 80
```

When we apply this Istio VirtualService to the cluster, the `plants` logs report errors:


```bash
writing a new plant to Firestore...
🚫 Failed adding plant: rpc error: code = Unavailable desc = all SubConns are in TransientFailure
```

And in the service graph, we can see that the `firestore` node has a purple `VirtualService` icon, meaning we've applied an Istio traffic rule against that service. Eventually, throughput to firestore will appear as `0` over the last minute, since we've redirected all outgoing connections to the database.

![kiali](/images/databases-kiali.png)

Note that you can also use Istio to manage traffic for databases *inside* the cluster, including Redis, SQL, and [MongoDB](https://istio.io/blog/2018/egress-mongo/). See the Istio docs to learn more.