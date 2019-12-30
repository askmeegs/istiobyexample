---
title: Secure Ingress
---

If you're running workloads in a Kubernetes cluster, it's likely that some need to be exposed outside of the cluster. The [Istio Ingress Gateway](/ingress) is a customizable proxy that can route inbound traffic for one or many backend hosts. But what about securing ingress traffic with HTTPS?

Istio supports TLS ingress by mounting certs and keys into the Ingress Gateway, allowing you to securely route inbound traffic to your in-cluster Services. When you set up secure ingress with Istio, the Ingress Gateway handles all TLS operations (handshake, certs/keys exchange), allowing you to decouple TLS from your application code. Further, using the Ingress Gateway for TLS traffic allows you to centralize and automate the management of certs and keys across your organization.

Istio supports securing the Ingress Gateway through two methods. The first is through [file mount](https://istio.io/docs/tasks/traffic-management/ingress/secure-ingress-mount/), where you generate certs and keys for the IngressGateway, then mount them manually into the IngressGateway as a Kubernetes Secret. The second way is through the [Secrets Discovery Service](https://istio.io/docs/tasks/traffic-management/ingress/secure-ingress-sds/) (SDS), an agent that runs in the IngressGateway pod, alongside the Istio proxy. The SDS agent  monitors the `istio-system` namespace for new secrets, and mounts them into the Gateway's proxy on your behalf. Like the file mount method, SDS supports both server-side and mutual TLS.

Let's see how to use the SDS method to configure the Ingress Gateway with mutual HTTPS authentication.

![](/images/secure-ingress-arch.png)

Here, a construction materials company called FooCorp runs one production Kubernetes cluster. One team, `ux`, runs a customer-facing web `frontend`. The other, `corp-services`, runs an internal-facing `inventory` for supply chain tracking. Both services have their own `foocorp` subdomain, and the security team has mandated that every service has its own certs and keys. Let's walk through the configuration of secure ingress on this cluster.

First, we'll install Istio, enabling the [global SDS ingress](https://istio.io/docs/reference/config/installation-options/#gateways-options option) option. When we enable this, the Istio `ingress-gateway` pod will have two containers, `istio-proxy` (Envoy) and `ingress-sds`, which is the Secrets Discovery agent:

```
istio-ingressgateway-6f7d65d984-m2zmn     2/2     Running     0          44s
```

Then we'll create two namespaces, `ux` and `corp-services`, and label both for Istio sidecar proxy injection. Next, we'll apply Deployments and Services for the `frontend` (`ux` namespace) and the `inventory` (`corp-services` namespace).

Now, we're ready to generate certs and keys for two domains, `frontend.foocorp.com` and `inventory.foocorp.com`. (Note: you don't need to purchase domain names to try this out - we'll test with the `host` header in a few steps.) We generate Kubernetes secrets from these keys:

```
kubectl create -n istio-system secret generic frontend-credential  \
--from-file=key=frontend.foocorp.com/3_application/private/frontend.foocorp.com.key.pem \
--from-file=cert=frontend.foocorp.com/3_application/certs/frontend.foocorp.com.cert.pem \
--from-file=cacert=frontend.foocorp.com/2_intermediate/certs/ca-chain.cert.pem

kubectl create -n istio-system secret generic inventory-credential  \
--from-file=key=inventory.foocorp.com/3_application/private/inventory.foocorp.com.key.pem \
--from-file=cert=inventory.foocorp.com/3_application/certs/inventory.foocorp.com.cert.pem \
--from-file=cacert=inventory.foocorp.com/2_intermediate/certs/ca-chain.cert.pem
```

Now, we're ready to expose `frontend` and `inventory` with Istio resources. First, create a `Gateway` resource, punching port `443` for HTTPS traffic. Note that `mode: MUTUAL` indicates that we're enforcing mutual TLS for inbound traffic. And for each service, we specify two different sets of credentials, corresponding to the Secrets we just created.

```YAML
apiVersion: networking.istio.io/v1alpha3
kind: Gateway
metadata:
  name: foocorp-gateway
  namespace: default
spec:
  selector:
    istio: ingressgateway # use istio default ingress gateway
  servers:
  - port:
      number: 443
      name: https-frontend
      protocol: HTTPS
    tls:
      mode: MUTUAL
      credentialName: "frontend-credential"
    hosts:
    - "frontend.foocorp.com"
  - port:
      number: 443
      name: https-inventory
      protocol: HTTPS
    tls:
      mode: MUTUAL
      credentialName: "inventory-credential"
    hosts:
    - "inventory.foocorp.com"
```

Next, create two Istio VirtualServices to handle routing from the Gateway. Since both services are mapped to port `443` in the Gateway, we'll use the `host` header (or a domain name) to specify the backend service we're requesting.

```YAML
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: frontend
spec:
  hosts:
  - "frontend.foocorp.com"
  gateways:
  - foocorp-gateway
  http:
  - match:
    - uri:
        exact: /
    route:
    - destination:
        host: frontend.ux.svc.cluster.local
        port:
          number: 80
---
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: inventory
spec:
  hosts:
  - "inventory.foocorp.com"
  gateways:
  - foocorp-gateway
  http:
  - match:
    - uri:
        exact: /
    route:
    - destination:
        host: inventory.corp-services.svc.cluster.local
        port:
          number: 80
```

Apply these YAML resources, then get the `istio-ingressgateway` pod logs for the `ingress-sds` container. You should see that when we applied a resource using specific credentials, the SDS agent mounted those certs and keys into the ingress proxy:

```bash
istio-ingressgateway-6f7d65d984-m2zmn ...
RESOURCE NAME:inventory-credential, EVENT: pushed key/cert pair to proxy
```

Now, we're ready to make requests to our two services from outside the cluster. Note that because we've configured mutual TLS, we have to specify `cert` and `key` in addition to `ca-cert`, in order for the server (the Ingress Gateway) to verify the identity of the client.

First, from any host outside the cluster, curl the frontend, with the frontend client keys:

```
$ curl -HHost:frontend.foocorp.com \
--resolve frontend.foocorp.com:$SECURE_INGRESS_PORT:$INGRESS_HOST \
--cacert frontend.foocorp.com/2_intermediate/certs/ca-chain.cert.pem \
--cert frontend.foocorp.com/4_client/certs/frontend.foocorp.com.cert.pem \
--key frontend.foocorp.com/4_client/private/frontend.foocorp.com.key.pem \
https://frontend.foocorp.com:$SECURE_INGRESS_PORT/

🏗 Welcome to FooCorp - Public Site
```

And the internal inventory, with the inventory keys:

```
$ curl -HHost:inventory.foocorp.com \
--resolve inventory.foocorp.com:$SECURE_INGRESS_PORT:$INGRESS_HOST \
--cacert inventory.foocorp.com/2_intermediate/certs/ca-chain.cert.pem \
--cert inventory.foocorp.com/4_client/certs/inventory.foocorp.com.cert.pem \
--key inventory.foocorp.com/4_client/private/inventory.foocorp.com.key.pem \
https://inventory.foocorp.com:$SECURE_INGRESS_PORT/

🛠 FooCorp - Inventory [INTERNAL]
```

What's actually happening here? Let's look at the inventory service, and walk through exactly how the Ingress Gateway authenticates the client.

![](/images/secure-ingress-auth-steps.png)

1. Client requests the host `https://inventory.foocorp.com:443`
2. DNS for `inventory.foocorp.com` resolves to the Istio Ingress Gateway's public IP, provisioned by default with a Kubernetes Service `type=LoadBalancer`. The Ingress Gateway presents its cert and key to the client.
3. Client verifies the Ingress Gateway's identity with the Certificate Authority (CA).
4. Client presents its cert and key to the Ingress Gateway.
5. Server (Ingress Gateway) verifies client's identity with the CA.
6. A secure connection is established between the client and the Ingress Gateway, and the Ingress Gateway forwards requests to the `inventory` Service.

🎊 We did it! From here, you can keep adding new services, and scale out the Ingress Gateway replicas to support a secure, centrally-managed ingress for your cluster.

**Learn More:**

- [Istio Ingress Gateway - Concepts](https://istio.io/docs/concepts/traffic-management/#gateways)
- [Istio SDS Ingress, server-side TLS only](https://istio.io/docs/tasks/traffic-management/ingress/secure-ingress-sds/#configure-a-tls-ingress-gateway-for-multiple-hosts)
- [Istio SDS Ingress, manual file-mount approach](https://istio.io/docs/tasks/traffic-management/ingress/secure-ingress-mount/#before-you-begin)
