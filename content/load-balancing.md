---
title: Load Balancing
publishDate: "2019-12-31"
categories: ["Traffic Management"]
---

Kubernetes supports load balancing for [inbound](https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer) traffic. But what about Kubernetes services inside the cluster?

When in-cluster services [communicate](https://kubernetes.io/docs/concepts/services-networking/#proxy-mode-iptables), a load balancer called kube-proxy [forwards](https://cloud.google.com/kubernetes-engine/docs/concepts/network-overview#services) requests to service pods at random. You can use Istio to add more complex load balancing methods, enabled by Envoy.

Envoy supports multiple load balancing methods, including [random](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/load_balancers#random), [round-robin](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/load_balancers#weighted-round-robin), and [least request](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/load_balancers#weighted-least-request).

Let's see how to use Istio to add [**least request**](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/upstream/load_balancing/load_balancers#weighted-least-request) load balancing for a service called `payments`, which processes all transactions for a web `frontend`. The payments service is backed by three pods.

In this least request algorithm, the client-side Envoy will first choose two instances at random. Then, it will forward the request to the instance with the fewest number of active requests, to help ensure even load balancing across all instances.

![load balancing](/images/lb-least-requests.png)

To enable this functionality, we create an Istio [DestinationRule](https://istio.io/docs/reference/config/networking/v1alpha3/destination-rule/) with a [`LEAST_CONN`](https://istio.io/docs/reference/config/networking/v1alpha3/destination-rule/#LoadBalancerSettings-SimpleLB) load balancer:

```YAML
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: payments-load-balancer
spec:
  host: payments.prod.svc.cluster.local
  trafficPolicy:
      loadBalancer:
        simple: LEAST_CONN
```

Check out the [Istio docs](https://istio.io/docs/concepts/traffic-management/#load-balancing-3-subsets) to see how to add multiple load balancing methods for a single host.