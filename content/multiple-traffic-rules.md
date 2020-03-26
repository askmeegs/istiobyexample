---
title: "Multiple Traffic Rules"
lastmod: "2019-12-31"
publishDate: "2019-12-31"
categories: ["Traffic Management"]
---

Istio supports lots of [traffic management](https://istio.io/docs/concepts/traffic-management/) use cases, from [redirects](https://istio.io/docs/reference/config/networking/virtual-service/#HTTPRewrite) and [traffic splitting](https://istio.io/docs/tasks/traffic-management/traffic-shifting/) to [mirroring](https://istio.io/docs/tasks/traffic-management/mirroring/) and [retry logic](https://istio.io/docs/concepts/traffic-management/#retries). If you've created an Istio [VirtualService](https://istio.io/docs/reference/config/networking/virtual-service/) to define one of these policies for a service, it's easy to add more traffic management rules to the same resource. This example demonstrates how to apply multiple traffic rules to one Kubernetes-based service.

Let's say that we're on the **frontend** engineering team for a newspaper's website. The user-facing frontend service relies on a backend called **articles**, which serves article content and metadata as JSON. But the articles team is refactoring the service in a new language, and they're frequently rolling out new changes. This has resulted in unexpected errors in the frontend, which relied on the behavior of the legacy articles service. To complicate things, the newspaper's blog, previously served by a separate **blog** service, was just incorporated into the articles service. Now, all blog posts are articles, served with the `/beta/blog` path.

In order to lock in the behavior of articles on behalf of the frontend, we'll create an Istio traffic policy for articles. The frontend's traffic requirements for articles include: returning a `no-cache` header for any `/breaking-news` article, rewriting `/blog` to `/beta/blog`, and enforcing a 2-second timeout on every request.

![](/images/multiple-functionality.png)

To get this aggregate functionality, we'll create one VirtualService for articles, with three `http` rules: [header modification](/response-headers) for `/breaking-news`, URL rewrite for `/blog`, and a default fallthrough to the articles service. All three rules have a 2-second timeout.

![](/images/multiple-vs.png)


```YAML
apiVersion: networking.istio.io/v1alpha3
kind: VirtualService
metadata:
  name: articles-vs
spec:
  hosts:
  - articles
  http:
  - match: # RULE 1 - BREAKING NEWS
    - uri:
        prefix: "/article/breaking-news"
    route:
    - destination:
        host: articles
      headers:
        response:
          add:
            no-cache: "true"
      timeout: 2s
  - match: # RULE 2 - BLOG URI REWRITE
    - uri:
        prefix: /blog
    rewrite:
      uri: /beta/blog
    route:
    - destination:
        host: articles
      timeout: 2s
  - route: # RULE 3 / DEFAULT - TIMEOUT
    - destination:
        host: articles
    timeout: 2s
    weight: 100
```

After applying this VirtualService to the cluster, we can exec into the frontend pod and access the articles service, to verify that the rules are taking effect. For instance, if we request a `/breaking-news` article, we'll see the `no-cache:true` header added to the response:

```bash
$ curl -v http://articles:80/article/breaking-news/2020/astrophysics-discovery

< HTTP/1.1 200 OK
< date: Wed, 15 Jan 2020 22:10:52 GMT
...
< no-cache: true
```

If we request a path starting with `/blog`, that gets rewritten to `/beta/blog`, and articles knows to serve the `/beta/blog` path:

```bash
$ curl http://articles:80/blog/2020/new-engineering-blog

{"id":91385,"title":"Welcome to the new News Blog!" ...
```

And finally, if we write in a 10-second sleep into the article service's application code, we can see our 2-second timeout in action:

```bash
$ curl  http://articles:80/

upstream request timeout
```

![](/images/multiple-fault.png)

**Note**: When you add multiple rules to one VirtualService, [the rules are evaluated **in order**](https://istio.io/docs/concepts/traffic-management/#routing-rule-precedence), from top to bottom. So if we add a [fault injection](https://istio.io/docs/tasks/traffic-management/fault-injection/#injecting-an-http-abort-fault) rule to the articles VirtualService to return HTTP status code `404: Not Found` on all requests, that rule will override the other three, taking down the entire articles service:

```bash
curl -v http://articles:80
...
* Connected to articles (10.0.21.31) port 80

< HTTP/1.1 404 Not Found
```

Because of this rule precedence for VirtualServices, it's important to validate and test complex Istio traffic policies. It's also a good idea to add a "fallthrough" or default rule, as we did above, to ensure that all requests for a specific host get routed.