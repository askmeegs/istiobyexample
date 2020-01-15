WORKDIR="`pwd`"
ISTIO_VERSION="${ISTIO_VERSION:-1.4.3}"
log "Downloading Istio ${ISTIO_VERSION}..."
curl -L https://git.io/getLatestIstio | ISTIO_VERSION=$ISTIO_VERSION sh -


# Prepare for install
kubectl label namespace default istio-injection=enabled
kubectl create namespace istio-system

kubectl create clusterrolebinding cluster-admin-binding \
    --clusterrole=cluster-admin \
    --user=$(gcloud config get-value core/account)

helm template ${WORKDIR}/istio-${ISTIO_VERSION}/install/kubernetes/helm/istio-init --name istio-init --namespace istio-system | kubectl apply -f -
sleep 20



# installs Istio with Envoy access logging enabled
helm template ${WORKDIR}/istio-${ISTIO_VERSION}/install/kubernetes/helm/istio --name istio --namespace istio-system \
--set prometheus.enabled=true \
--set tracing.enabled=true \
--set kiali.enabled=true --set kiali.createDemoSecret=true \
--set "kiali.dashboard.jaegerURL=http://jaeger-query:16686" \
--set "kiali.dashboard.grafanaURL=http://grafana:3000" \
--set grafana.enabled=true \
--set mixer.policy.enabled=false \
--set global.proxy.accessLogFile="/dev/stdout" >> istio.yaml

# install istio
kubectl apply -f istio.yaml