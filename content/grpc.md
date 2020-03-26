---
title: "gRPC"
lastmod: "2019-12-31"
publishDate: "2019-12-31"
categories: ["Traffic Management"]
---

[gRPC](https://grpc.io/) is a communication protocol for services, built on [HTTP/2](https://www.cncf.io/blog/2018/08/31/grpc-on-http-2-engineering-a-robust-high-performance-protocol/). Unlike REST over HTTP/1, which is based on [resources](https://en.wikipedia.org/wiki/Representational_state_transfer), gRPC is based on [Service Definitions](https://grpc.io/docs/guides/concepts/). You specify service definitions in a format called [protocol buffers](https://developers.google.com/protocol-buffers/) ("proto"), which can be serialized into an small binary format for transmission.

With gRPC, you can generate boilerplate code from `.proto` files into [multiple programming languages](https://grpc.io/docs/quickstart/), making gRPC an ideal choice for polyglot microservices.

While gRPC supports some networking use cases like [TLS](https://grpc.io/docs/guides/auth/) and [client-side load balancing](https://grpc.io/blog/loadbalancing/), adding Istio to a gRPC architecture can be useful for collecting telemetry, adding traffic rules, and setting [RPC-level authorization](https://istio.io/blog/2018/istio-authorization/#rpc-level-authorization). Istio can also provide a useful management layer if your traffic is a mix of HTTP, TCP, gRPC, and database protocols, because you can use the same Istio APIs for all traffic types.

[Istio](https://istio.io/about/feature-stages/#traffic-management) and its data plane proxy, [Envoy](https://www.envoyproxy.io/docs/envoy/latest/intro/arch_overview/other_protocols/grpc#arch-overview-grpc), both  support gRPC. Let's see how to manage gRPC traffic with Istio.

![grpc](/images/grpc.png)

Here, we're running two gRPC Services, `client` and `server`. `client` makes an RPC call to the `server`'s `/SayHello` function every 2 seconds.


Adding Istio to gRPC Kubernetes services has one pre-requisite: [labeling](https://istio.io/docs/setup/kubernetes/additional-setup/requirements/) your Kubernetes Service ports. The server's port is labeled as follows:

```YAML
apiVersion: v1
kind: Service
metadata:
  name: server
spec:
  selector:
    app: server
  type: ClusterIP
  ports:
  - name: grpc # important!
    protocol: TCP
    port: 8080
```

Once we deploy the app, we can see this traffic between client and server in a [service graph](https://www.kiali.io/):

![kiali](/images/grpc-kiali.png)

We can also view the server's gRPC traffic metrics in Grafana:

![](/images/grpc-server-healthy.png)

Then, we can apply an Istio traffic rule to inject a 10-second delay [fault](https://istio.io/docs/tasks/traffic-management/fault-injection/) into `server`. You might apply this rule in a chaos testing scenario, to test the resiliency of this application.

```YAML
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: server-fault
spec:
  hosts:
  - server
  http:
  - fault:
      delay:
        percentage:
          value: 100.0
        fixedDelay: 10s
    route:
    - destination:
        host: server
        subset: v1
```

This causes the client RPC to time out (`Outgoing Request Duration`):

![](/images/grpc-grafana-client-fault-inject.png)


To learn more about gRPC and Istio:
- [Istio docs - Traffic Management](https://istio.io/docs/concepts/traffic-management/#traffic-routing-and-configuration)
- [Blog post - gRPC + Istio + Cloud Internal Load Balancer](https://cloud.google.com/solutions/using-istio-for-internal-load-balancing-of-grpc-services)