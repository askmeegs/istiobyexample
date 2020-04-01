---
title: 持ち込みPrometheus
publishDate: "2019-12-31"
categories: ["Observability"]
---

[Prometheus](https://prometheus.io/docs/introduction/overview/)は、オープンソースの監視ツールです。デフォルトでは、PrometheusはIstioと一緒にインストールされるため、GrafanaとKialiを使用して、IstioコントロールプレーンとEnvoyを利用したワークロードの両方のメトリクスを表示できます。

しかし、すでにクラスタでPrometheusを実行している場合、またはIstioのPrometheusインストールに追加のカスタマイズを行う場合（たとえば、Istioの[Slack通知](https://prometheus.io/docs/alerting/notification_examples/#customizing-slack-notifications)を追加する場合）はどうなりますか？

心配要りません。 3つの簡単な手順で、ご自身で管理しているPrometheusをIstioに持ち込むことができます。

まず、Prometheus構成を更新します。 Prometheusは、targets が /memetrics エンドポイントを表し、Prometheusサーバーによって取り込まれた[スクレイプ構成モデル](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#scrape_config)に依存しています。

[各Istioコンポーネント](https://istio.io/docs/tasks/observability/metrics/querying-metrics/)の targets を追加します。これらは、[Kubernetes APIサーバーを介して](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#kubernetes_sd_config)取得されます。たとえば、Istioの[Pilot](https://istio.io/docs/concepts/traffic-management/#pilot)コンポーネントの構成は次のとおりです。

```YAML
    - job_name: 'pilot'
      kubernetes_sd_configs:
      - role: endpoints
        namespaces:
          names:
          - istio-system

      relabel_configs:
      - source_labels: [__meta_kubernetes_service_name, __meta_kubernetes_endpoint_port_name]
        action: keep
        regex: istio-pilot;http-monitoring
```

完全な例として、[configmap.yaml](https://github.com/askmeegs/istiobyexample/blob/888a7b5c573c9ba6bf2c0e046e44bf4f8d8d2506/content/blog/prometheus/configmap.yaml)をご覧ください。

次に、PrometheusのDeploymentを更新して、Istioの証明書をPrometheusにマウントします。これにより、相互TLSが有効になっている場合、PrometheusはIstioワークロードをスクレイピングできます。これを行うには、 istio.default シークレットをPrometheusのDeployment.yamlに設定します。

```YAML
    volumes:
    - name: config-volume
        configMap:
        name: prometheus
    - name: istio-certs
        secret:
            defaultMode: 420
            optional: true
            secretName: istio.default
```

完全な例として、[deployment.yaml](https://github.com/askmeegs/istiobyexample/blob/888a7b5c573c9ba6bf2c0e046e44bf4f8d8d2506/content/blog/prometheus/deployment.yaml)をご覧ください。

この新しい構成でPrometheusをデプロイすると、DeploymentとServiceが別のモニタリング名前空間で実行されます。

```bash
$ kubectl get service -n monitoring

NAME         TYPE           CLUSTER-IP   EXTERNAL-IP      PORT(S)          AGE
prometheus   LoadBalancer   10.0.3.155   <IP>             9090:32352/TCP   21m
```

最後に、カスタムのPrometheusアドレスを使用するようにIstioの構成を更新します。 [Istioインストールオプション](https://istio.io/docs/reference/config/installation-options/#grafana-options)を使用した Helm template の例を次に示します。

```bash
helm template install/kubernetes/helm/istio --name istio --namespace istio-system \
--set prometheus.enabled=false \
--set kiali.enabled=true --set kiali.createDemoSecret=true \
--set kiali.prometheusAddr="http://prometheus.monitoring.svc.cluster.local:9090" \
--set "kiali.dashboard.jaegerURL=http://jaeger-query:16686" \
--set "kiali.dashboard.grafanaURL=http://grafana:3000" \
--set grafana.enabled=true \
--set grafana.datasources.datasources.datasources.url="http://prometheus.monitoring.svc.cluster.local:9090"  > istio.yaml
```

IstioとPrometheusの両方がインストールされ、Istioを利用したワークロードをクラスターにデプロイすると、PrometheusがIstioターゲットを正常にスクレイピングしていることがわかります。

![](/images/prometheus.png)

Grafanaはサービスレベルの指標を取得できます。

![](/images/prom-grafana.png)

そして、Kialiはサービスグラフを表示できます。

![](/images/prom-kiali.png)
