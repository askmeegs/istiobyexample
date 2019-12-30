---
title: Authorization
---

Authentication means verifying the identity of a client. **Authorization**, on the other hand, verifies the permissions of that client, or: "can this service do what they're asking to do?"

Kubernetes supports authorization through **role-based access control** ([RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#api-overview)). In this model, you can map a **role** (a set of permissions) to a **role binding** (who exactly has these permissions?).

[Istio uses RBAC](https://istio.io/docs/concepts/security/#authorization) to allow you to set authorization policies between services in your cluster. Let's see how.

Here, the _ShoeStore_ application is deployed to the `default` Kubernetes namespace. There are three services, and all speak plain HTTP:
1. `shoes`: exposes an API for all the shoes in the store
2. `users`: stores purchase history
3. `inventory`: loads new shoe models into `shoes`.

![shoes rbac](/images/rbac.png)

We want to authorize `inventory` to be able to `POST` new inventory to `shoes`, but not be able to access `users` at all. To do this, we will need three Istio resources.

First, we need to enable authorization for the `default` Kubernetes namespace:

```YAML
apiVersion: "rbac.istio.io/v1alpha1"
kind: ClusterRbacConfig
metadata:
  name: default
spec:
  mode: 'ON_WITH_INCLUSION'
  inclusion:
    namespaces: ["default"]
```

Second, we'll define our `ServiceRole`, the abstract set of permissions to `POST` data to the `shoes` service:

```YAML
apiVersion: "rbac.istio.io/v1alpha1"
kind: ServiceRole
metadata:
  name: shoes-user
  namespace: default
spec:
  rules:
  - services: ["shoes.default.svc.cluster.local"]
    methods: ["POST"]
```

Lastly, we'll bind this abstract `shoes-user` role to a concrete set of `subjects` -- in this case, any request that has the `hello:world` HTTP header.

```YAML
apiVersion: "rbac.istio.io/v1alpha1"
kind: ServiceRoleBinding
metadata:
  name: shoes-user-binding
  namespace: default
spec:
  subjects:
  - properties:
      request.headers[hello]: "world"
  roleRef:
    kind: ServiceRole
    name: "shoes-user"
```

From here, the `inventory` service can only make requests to `shoes` with the `hello:world` header, and the `users` service is still locked down. This is because RBAC is on, but we haven't assigned any `ServiceRoles` to access the `users` service.

To learn more:

- [Istio Authorization with Pod service accounts](https://istio.io/docs/tasks/security/authz-http/#step-2-allowing-access-to-the-details-and-reviews-services) (requires mutual TLS)
- [End-user authorization](https://istio.io/blog/2018/istio-authorization/#using-authenticated-client-identities) through JWT claims
- [Namespace-level authorization](https://istio.io/docs/tasks/security/authz-http/#enforcing-namespace-level-access-control)