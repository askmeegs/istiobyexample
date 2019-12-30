---
title: Authorization
---

Authentication means verifying the identity of a client. **Authorization**, on the other hand, verifies the permissions of that client, or: "can this service do what they're asking to do?"

While all requests in an Istio mesh are allowed by default, [Istio provides](https://istio.io/docs/concepts/security/#authorization) an [`AuthorizationPolicy` resource](https://istio.io/docs/reference/config/security/authorization-policy/) to define authorization rules for your services. Istio translates this resource into Envoy-readable config, and mounts that config into the Istio sidecar proxies. From here, authorization policy checks are performed by the sidecar proxies. Let's see how it works.

![shoes rbac](/images/rbac.png)

Here, the _ShoeStore_ application is deployed to the `default` Kubernetes namespace. There are three services, and all speak plain HTTP:
1. `shoes`: exposes an API for all the shoes in the store
2. `users`: stores purchase history
3. `inventory`: loads new shoe models into `shoes`.


We want to authorize `inventory` to be able to `POST` new inventory to `shoes`, but not be able to access `users` at all. To do this, we will create two `AuthorizationPolicies`: one for `shoes`, and one for `users`.

```YAML

```


To learn more:

- [Istio Authorization with Pod service accounts](https://istio.io/docs/tasks/security/authz-http/#step-2-allowing-access-to-the-details-and-reviews-services) (requires mutual TLS)
- [End-user authorization](https://istio.io/blog/2018/istio-authorization/#using-authenticated-client-identities) through JWT claims
- [Namespace-level authorization](https://istio.io/docs/tasks/security/authz-http/#enforcing-namespace-level-access-control)