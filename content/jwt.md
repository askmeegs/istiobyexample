---
title: JWT
publishDate: "2019-12-31"
categories: ["Security"]
---

A [JSON Web Token](https://jwt.io/introduction/) (JWT) is a type of authentication token used to identify a user to a server application. JWTs contain information about the client caller, and can be used as part of a client session architecture. A [JSON Web Key Set](https://auth0.com/docs/jwks) (JWKS) contains the cryptographic keys used to verify incoming JWTs.

You can use Istio's [Authentication API](https://istio.io/docs/reference/config/istio.authentication.v1alpha1/#Jwt) to [configure JWT policies](https://istio.io/docs/concepts/security/#origin-authentication) for your services.

![jwt](/images/jwt.png)

In this example, we require a JWT for all routes in the `frontend` service except for the home page (`/`) and the pod health check (`/_healthz`).

In the Istio policy, we specify the path to a test public key (`jwksUri`), which will be mounted into the frontend's sidecar proxy. All unauthenticated requests will receive a `401 - Unauthorized` status from Envoy.

```YAML
apiVersion: "authentication.istio.io/v1alpha1"
kind: "Policy"
metadata:
  name: "frontend-jwt"
spec:
  targets:
  - name: frontend
  origins:
  - jwt:
      issuer: "testing@secure.istio.io"
      jwksUri: "https://raw.githubusercontent.com/istio/istio/release-1.2/security/tools/jwt/samples/jwks.json"
      trigger_rules:
      - excluded_paths:
        - exact: /_healthz
        - exact: /
  principalBinding: USE_ORIGIN
```

To learn more and try interactive examples, see the [Istio docs](https://istio.io/docs/tasks/security/authn-policy/#end-user-authentication) and the [istio-samples repo](https://github.com/GoogleCloudPlatform/istio-samples/tree/master/security-intro#add-end-user-jwt-authentication).
