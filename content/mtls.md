---
title: Mutual TLS
---

A microservices architecture means more requests on the network, and more opportunities for malicious parties to intercept traffic. [Mutual TLS](https://en.wikipedia.org/wiki/Mutual_authentication) (mTLS) authentication is a way to encrypt services traffic using [certificates](https://www.internetsociety.org/deploy360/tls/basics/).

With Istio, you can [enforce mutual TLS automatically](https://istio.io/docs/concepts/security/#authentication-policies), outside of your application code, with a single YAML file. This works because the Istio control plane mounts client certificates into the sidecar proxies for you, so that pods can authenticate with each other.

Let's enable mutual TLS for the entire service mesh, including the two services (`client` and `server`) pictured below.

![Diagram](/images/mtls.png)

Starting in Istio 1.5, Istio uses [automatic mutual TLS](https://istio.io/pt-br/docs/tasks/security/authentication/auto-mtls/). This means that while services accept both plain-text and TLS traffic, by default, services will *send* TLS requests within the cluster. This means that the client-to-server above will already be encrypted with the default Istio install. (But HTTP will still work.)

We can test this by sending a plain HTTP request from the `client` pod's Istio proxy container to the `server`. The request succeeds:

```
$ curl http://server:80

Hello World! /$
```

To enforce mTLS such that *only* TLS traffic is accepted by all services in all Istio-injected namespacess, we'll apply the following [Istio `PeerAuthentication` resource](https://istio.io/docs/reference/config/security/peer_authentication/):

```YAML
apiVersion: "security.istio.io/v1beta1"
kind: "PeerAuthentication"
metadata:
  name: "default"
  namespace: "istio-system"
spec:
  mtls:
    mode: STRICT
```

Now let's try to send another request between `client` and `server`:

```
$ curl http://server:80
curl: (56) Recv failure: Connection reset by peer
```

If we view the Kiali service graph for this mesh, we can see a lock icon indicating that traffic is encrypted between `client` and `server`:

![](/images/mtls-kiali.png)


**Authentication Flow:**

1. `client` application container sends a plain-text HTTP request to `server`.
2. `client` proxy container intercepts the outbound request.
3. `client` proxy performs a TLS [handshake](https://www.ibm.com/support/knowledgecenter/en/SSFKSJ_7.1.0/com.ibm.mq.doc/sy10660_.htm) with the server-side proxy. This handshake includes an exchange of certificates. These certs are pre-loaded into the proxy containers by Istio.
4. `client` proxy performs a [secure naming](https://istio.io/docs/concepts/security/#secure-naming) check on the server's certificate, verifying that an authorized identity is running the `server`.
5. `client` and `server` establish a mutual TLS connection, and the `server` proxy forwards the request to the `server` application container.


**Learn more**:

- [YAML resources used for this example (Github)](https://github.com/askmeegs/istiobyexample/tree/master/yaml/mtls)
- [Istio Docs - Authentication](https://istio.io/docs/concepts/security/#authentication)
- [Samples: Authentication](https://github.com/GoogleCloudPlatform/istio-samples/tree/77fb1dfb690d28517e410df2911e255d54e3450e/security-intro#authentication)
- [Task - Authentication Policies](https://istio.io/docs/tasks/security/authentication/authn-policy/)
- [Task - Mutual TLS Deep-Dive](https://istio.io/docs/tasks/security/mutual-tls/)
- [FAQ - mTLS](https://istio.io/faq/security/#enabling-disabling-mtls)



