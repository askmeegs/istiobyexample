---
title: Traffic Mirroring
---

Testing a service in production is important to help ensure reliability. Sending live production traffic to a new version of a service can help reveal bugs that went untested during continuous integration and functional tests.

Using Istio, you can use [**traffic mirroring**](https://istio.io/docs/tasks/traffic-management/mirroring/) to duplicate traffic to another service. You can incorporate a traffic mirroring rule as part of a [canary deployment](https://istiobyexample.dev/canary) pipeline, allowing you to analyze a service's behavior before sending live traffic to it.

In this example, we have deployed a video processing pipeline to Kubernetes. The `render` service has a dependency on the `encode-prod` service, and we want to roll out a new version of `encode`, `encode-test`.

![traffic mirroring](/images/traffic-mirror.png)

We can use an Istio `VirtualService` to mirror all `encode-prod` traffic to `encode-test`. The client side Envoy proxy for `render` will then send requests to both `encode-prod` (request path) and `encode-test` (mirrored path). `prod` and `test` are two subsets of the `encode` Kubernetes service, specified in an Istio `DestinationRule`.

**Note**: `render` [will not wait](https://www.envoyproxy.io/docs/envoy/latest/api-v2/api/v2/route/route.proto#route-routeaction-requestmirrorpolicy) to receive responses from `encode-test` — mirrored requests are "fire and forget."

```YAML
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: encode-mirror
spec:
  hosts:
    - encode
  http:
  - route:
    - destination:
        host: encode
        subset: prod
      weight: 100
    mirror:
      host: encode
      subset: test
```

```YAML
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: encode
spec:
  host: encode
  subsets:
  - name: prod
    labels:
      version: prod
  - name: test
    labels:
      version: test
```