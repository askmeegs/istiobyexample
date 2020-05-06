---
title: JWT
---

A [JSON Web Token](https://jwt.io/introduction/) (JWT) is a type of authentication token used to identify a user to a server application. JWTs contain information about the client caller, and can be used as part of a client session architecture. A [JSON Web Key Set](https://auth0.com/docs/jwks) (JWKS) contains the cryptographic keys used to verify incoming JWTs.

You can use Istio's [RequestAuthentication](https://istio.io/docs/reference/config/security/request_authentication/) resource to [configure JWT policies](https://istio.io/docs/concepts/security/#authentication-architecturen) for your services.

![jwt](/images/jwt.png)

By default, we can reach the frontend service through a `curl` request to the Istio IngressGateway's public IP:

```
$ curl ${INGRESS_IP}

Hello World! /
```

Now, let's require a JWT for all requests to the `frontend` service.
To do this, we'll need two Istio resources. The first is the `RequestAuthentication` policy that validates incoming tokens:

```YAML
apiVersion: security.istio.io/v1beta1
kind: RequestAuthentication
metadata:
 name: frontend
 namespace: default
spec:
  selector:
    matchLabels:
      app: frontend
  jwtRules:
  - issuer: "testing@secure.istio.io"
    jwksUri: "https://raw.githubusercontent.com/istio/istio/release-1.5/security/tools/jwt/samples/jwks.json"
```

The second resource is an `AuthorizationPolicy`, which ensures that all requests have a JWT - and rejects requests that do not, returning a `403` error.

```YAML
apiVersion: security.istio.io/v1beta1
kind: AuthorizationPolicy
metadata:
  name: require-jwt
  namespace: default
spec:
  selector:
    matchLabels:
      app: frontend
  action: ALLOW
  rules:
  - from:
    - source:
       requestPrincipals: ["testing@secure.istio.io/testing@secure.istio.io"]
```

Once we apply these resources, we can curl the Istio ingress gateway without a JWT, and see that the `AuthorizationPolicy` is rejecting our request because we did not supply a token:

```
$ curl ${INGRESS_IP}

RBAC: access denied
```

Then, we can curl again, this time with an invalid JWT. We receive an authentication error:
```
$ curl --header "Authorization: Bearer ${INVALID_JWT}" ${INGRESS_IP}

Jwt issuer is not configured
```

Finally, if we curl with a valid JWT, we can successfully reach the frontend via the IngressGateway:

```
$ curl --header "Authorization: Bearer ${VALID_JWT}" ${INGRESS_IP}

Hello World! /
```

To learn more and try interactive examples, see the [Istio docs](https://istio.io/docs/tasks/security/authorization/authz-jwt/) and the [istio-samples repo](https://github.com/GoogleCloudPlatform/istio-samples/tree/master/security-intro).



