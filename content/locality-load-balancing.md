---
title: Locality Load Balancing
lastmod: "2019-12-31"
publishDate: "2019-12-31"
categories: ["Traffic Management"]
---

If you're running a high-scale, global application, you might be running services in multiple regions. If you have multiple replicas of the same service, you may want to direct client requests to the closest server, in order to minimize latency. You might also want a way to handle failover if one region goes down, and direct traffic to the closest available service.

Istio can help you automatically handle regional traffic using a feature called **locality load balancing.** Let's see how.

![default](/images/loc-default.png)

Here, we have two Kubernetes clusters running in two different cloud regions, `us-central` and `us-east`.
The Istio control plane is running in `us-east`, and we have set up [single control plane](https://github.com/GoogleCloudPlatform/istio-samples/tree/191859c03e73da7e98d451c967cefe24101d1933/multicluster-gke/single-control-plane#demo-multicluster-istio--single-control-plane) Istio multicluster, so that services running in both clusters can reach each other.

When we started both clusters, the cloud provider added region-specific `failure-domain` labels to the Kubernetes nodes:

```
failure-domain.beta.kubernetes.io/region: us-central1
failure-domain.beta.kubernetes.io/zone: us-central1-b
```

Istio will populate requests with these locality labels, allowing Istio to redirect requests to the closest available region.

Both clusters are running an Istio-injected service called `echo`, which prints its location when accessed on port `80`. The central cluster is also running a `loadgen` service that calls `echo.default.svc.cluster.local:80` every second.

By default, the Kubernetes Service behavior is round-robin, between the two `echo` servers on both clusters:

```
$ 🌊 Hello World! - EAST
$ ✨ Hello World! - CENTRAL
$ 🌊 Hello World! - EAST
$ ✨ Hello World! - CENTRAL
```

We can enable locality load balancing by adding an [outlier detection](https://istio.io/docs/reference/config/networking/v1alpha3/destination-rule/#OutlierDetection) Istio DestinationRule on the `east` cluster:

```YAML
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: echo-outlier-detection
spec:
  host: echo.default.svc.cluster.local
  trafficPolicy:
    connectionPool:
      tcp:
        maxConnections: 1000
      http:
        http2MaxRequests: 1000
        maxRequestsPerConnection: 10
    outlierDetection:
      consecutiveErrors: 7
      interval: 30s
      baseEjectionTime: 30s
```

Now, all `loadgen` requests are routed to the closest instance of `echo`, running in `us-central`:

```
$ ✨ Hello World! - CENTRAL
$ ✨ Hello World! - CENTRAL
$ ✨ Hello World! - CENTRAL
```

![locality](/images/loc-locality.png)

If we delete the `echo` Deployment running in `us-central`, Istio will redirect `loadgen` requests to the `echo` Pod running in `us-east`:

```
$ 🌊 Hello World! - EAST
$ 🌊 Hello World! - EAST
$ 🌊 Hello World! - EAST
```

![failover](/images/loc-failover.png)

We can also add a percentage-based load balancing rule for mesh-wide traffic, in the global Istio installation settings:

```
    localityLbSetting:
      distribute:
      - from: us-central1/*
        to:
          us-central1/*: 20
          us-east1/*: 80
```

Now, all services running in both clusters will share requests 80/20, between `us-east` and `us-central`. No VirtualServices are needed.

```
$ 🌊 Hello World! - EAST
$ 🌊 Hello World! - EAST
$ 🌊 Hello World! - EAST
$ 🌊 Hello World! - EAST
$ ✨ Hello World! - CENTRAL
```

![split](/images/loc-splittraffic.png)


To learn more about Locality-based load balancing with Istio, see the [Istio documentation](https://istio.io/docs/ops/traffic-management/locality-load-balancing/).