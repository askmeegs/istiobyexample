---
title: External Services
---

A [service mesh](https://istio.io/docs/concepts/what-is-istio/#what-is-a-service-mesh) often spans one environment— for instance, one Kubernetes cluster. And together, all the connected services in that environment form the management domain of that mesh, from which you can view metrics and set policies.

But what if you are also running services *outside* the cluster, or you depend on external APIs?

Have no fear. Istio provides a resource called a [`ServiceEntry`](https://istio.io/docs/concepts/traffic-management/#service-entries) that lets you logically bring external services into your mesh -- even services you don't own.

When you create a ServiceEntry for an external hostname, you can view metrics and traces reaching all the way to that external service. You can even set traffic policies like [retry logic](/retry/) on those external services. Adding `ServiceEntries` effectively expands the reach of your Istio management domain. Let's see an example.

![external currency service](/images/ext-currency.png)

Here, we're running a global store website with a `currency` service, responsible for converting product prices based on a user's locality. We rely on an third-party currency conversion API, the [European Central Bank](https://www.ecb.europa.eu/stats/policy_and_exchange_rates/euro_reference_exchange_rates/html/index.en.html), to provide realtime exchange rates.

We want to set a 3-second timeout on all calls from services inside the mesh to this external API. We'll need two Istio resources to do this.

First, a `ServiceEntry`, which logically adds the European Central Bank's hostname, `ecb.europa.eu`, to the mesh:

```YAML
apiVersion: networking.istio.io/v1alpha3
kind: ServiceEntry
metadata:
  name: currency-api
spec:
  hosts:
  - www.ecb.europa.eu
  ports:
  - number: 80
    name: http
    protocol: HTTP
  - number: 443
    name: https
    protocol: HTTPS
```

Second, a `VirtualService` traffic rule, to set a timeout for calls to the API:

```YAML
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: currency-timeout
spec:
  hosts:
    - www.ecb.europa.eu
  http:
  - timeout: 3s
    route:
      - destination:
          host: www.ecb.europa.eu
        weight: 100
```

Once we create a ServiceEntry for the currency API, we can automatically see `ecb.europa.eu` in our [Kiali service graph](https://istio.io/docs/tasks/telemetry/kiali/) (and instantly know *who*'s calling it):

![service graph](/images/ext-servicegraph.png)


And we also get an automatic [Grafana dashboard](https://istio.io/docs/tasks/telemetry/metrics/using-istio-dashboard/) for this external service, to view data like response codes and latency:

![grafana](/images/ext-grafana.png)

[See the Istio docs](https://istio.io/docs/tasks/traffic-management/egress/egress-control/#manage-traffic-to-external-services) to learn more about managing and securing external services.