---
title: Authorization
publishDate: "2019-12-31"
categories: ["Security"]
---

Authentication means verifying the identity of a client. **Authorization**, on the other hand, verifies the permissions of that client, or: "can this service do what they're asking to do?"

While all requests in an Istio mesh are allowed by default, [Istio provides](https://istio.io/docs/concepts/security/#authorization) an [`AuthorizationPolicy` resource](https://istio.io/docs/reference/config/security/authorization-policy/) that allows you to define granular policies for your workloads. Istio translates your AuthorizationPolicies into Envoy-readable config, then mounts that config into the Istio sidecar proxies. From there, authorization policy checks are performed by the sidecar proxies. Let's see how it works.

![shoes rbac](/images/rbac.png)

Here, the _ShoeStore_ application is deployed to the `default` Kubernetes namespace. There are three HTTP workloads, each defined with their own Kubernetes Deployment, Service, and ServiceAccount.

1. **shoes**: exposes an API for all the shoes in the store
2. **users**: stores purchase history
3. **inventory**: loads new shoe models into shoes.

We want to authorize the inventory service to be able to `POST` data to the shoes services, and then lock down all access to the users service. To do this, we will create two `AuthorizationPolicies`: one for `shoes`, and one for `users`.

Before deploying any policies, we can access both shoes and users from inside the inventory service's application container.

```bash
$ curl -X GET shoes
🥾 Shoes service
```

```bash
$ curl -x GET users
👥 Users service
```

First, let's create an AuthorizationPolicy for `shoes`:

```YAML
apiVersion: "security.istio.io/v1beta1"
kind: "AuthorizationPolicy"
metadata:
  name: "shoes-writer"
  namespace: default
spec:
  selector:
    matchLabels:
      app: shoes
  rules:
  - from:
    - source:
        principals: ["cluster.local/ns/default/sa/inventory-sa"]
    to:
    - operation:
        methods: ["POST"]
```

In this policy:
- The `selector` on `shoes` means we're enforcing any Deployment labeled with `app:shoes`.
- The `source` workload we're allowing has the `inventory-sa` identity. In a Kubernetes environment, this means that only pods with the `inventory-sa` [Service Account](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/) can access shoes. (**Note**: You must [enable Istio mutual TLS authentication](https://istio.io/docs/tasks/security/authentication/authn-policy/#globally-enabling-istio-mutual-tls) to use service account-based authorization policies. Mutual TLS is what allows the workload's [service account](https://istio.io/docs/concepts/security/#istio-security-vs-spiffe) certificates to be passed in the request.)
- The only whitelisted HTTP operation is `POST`, meaning that other HTTP operations, like `GET`, will be denied.



Once we apply the `shoes-writer` policy, we can successfully `POST` from inventory:

```
$ curl -X POST shoes
🥾 Shoes service
```

But `GET` requests from `inventory` are denied:

```
$ curl -X GET shoes
RBAC: access denied
```

And if we try to `POST` from a workload other than `inventory`, for instance, from `users`, the request will be denied:

```
$ curl -X POST shoes
RBAC: access denied
```

Next, let's create a "deny-all" policy for the users service:

```YAML
apiVersion: "security.istio.io/v1beta1"
kind: "AuthorizationPolicy"
metadata:
  name: "users-deny-all"
  namespace: default
spec:
  selector:
    matchLabels:
      app: users
```

Note that there are no `rules` for this service, just a `matchLabels` for our users Deployment. Also note that [the difference](https://istio.io/docs/concepts/security/#allow-all-and-deny-all) between a deny-all and an allow-all AuthorizationPolicy is subtle: in an allow-all policy, you would specify `rules: {}`.

Once we apply this resource, we are no longer able to access users from any of our services:

```
$ curl users
RBAC: access denied
```

To learn more:

- [Manifests for this example](https://github.com/askmeegs/istiobyexample/tree/master/yaml/authorization)
- [Istio blog - Introducing the Istio v1beta1 Authorization Policy](https://istio.io/blog/2019/v1beta1-authorization-policy/)
- [Istio docs - Authorization concepts](https://istio.io/docs/concepts/security/#authorization)
- [Istio docs - Authorization task](https://istio.io/docs/tasks/security/authorization/authz-http/)
- [Istio Samples - Introduction to Istio Security](https://github.com/GoogleCloudPlatform/istio-samples/tree/6fa69cf46424c055535ddbdc22f715e866c4d179/security-intro#demo-introduction-to-istio-security)
