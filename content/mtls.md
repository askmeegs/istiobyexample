---
title: Mutual TLS
publishDate: "2019-12-31"
categories: ["Security"]
---

A microservices architecture means more requests on the network, and more opportunities for malicious parties to intercept traffic. [Mutual TLS](https://en.wikipedia.org/wiki/Mutual_authentication) (mTLS) authentication is a way to encrypt services traffic using [certificates](https://www.internetsociety.org/deploy360/tls/basics/).

With Istio, you can automate the enforcement of mTLS across all services. Below, we enable mTLS for the entire mesh. Two pods in the cluster, `client` and `server`, are shown establishing a secure connection with the mTLS policy in place.

![Diagram](/images/mtls.png)


```YAML
apiVersion: authentication.istio.io/v1alpha1
kind: MeshPolicy
metadata:
  name: default
spec:
  peers:
  - mtls:
      mode: STRICT
```

```YAML
apiVersion: networking.istio.io/v1alpha3
kind: DestinationRule
metadata:
  name: default
  namespace: istio-system
spec:
  host: "*.local"
  trafficPolicy:
    tls:
      mode: ISTIO_MUTUAL
```

Here, a `MeshPolicy` enforces TLS for all services *receiving* requests (server-side), and a `DestinationRule` enforces TLS for all services *sending* requests (client-side), resulting in mutual ("both") TLS.

**Authentication Flow:**

1. `client` application container sends a plain-text HTTP request to `server`.
2. `client` proxy container intercepts the outbound request.
3. `client` proxy performs a TLS [handshake](https://www.ibm.com/support/knowledgecenter/en/SSFKSJ_7.1.0/com.ibm.mq.doc/sy10660_.htm) with the server-side proxy. This handshake includes an exchange of certificates. These certs are pre-loaded into the proxy containers by Istio.
4. `client` proxy performs a [secure naming](https://istio.io/docs/concepts/security/#secure-naming) check on the server's certificate, verifying that an authorized identity is running the `server`.
5. `client` and `server` establish a mutual TLS connection, and the `server` proxy forwards the request to the `server` application container.



**Learn more**:

- [Istio Docs - Authentication](https://istio.io/docs/concepts/security/#authentication)
- [Samples: Authentication](https://github.com/GoogleCloudPlatform/istio-samples/tree/77fb1dfb690d28517e410df2911e255d54e3450e/security-intro#authentication)
- [Task - Authentication Policies](https://istio.io/docs/tasks/security/authn-policy/)
- [Task - Mutual TLS Deep-Dive](https://istio.io/docs/tasks/security/mutual-tls/)
- [FAQ - mTLS](https://istio.io/faq/security/#enabling-disabling-mtls)
