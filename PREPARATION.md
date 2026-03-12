- have kind installed
- create a cluster: `kind create cluster --config cluster.kind`
- have cloud-provider-kind installed: https://kind.sigs.k8s.io/docs/user/loadbalancer/
- run `cloud-provider-kind` in a separate terminal and keep it running while using the cluster
- have helm-diff installed: `helm plugin install https://github.com/databus23/helm-diff`
- have nginx-ingress installed: `kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml`
- make sure the service for nginx-ingress has an external IP
- update `/etc/hosts` with the current ingress load balancer IP:
```
./scripts/print-kind-hosts.sh
```
- apply all: `helmfile apply --skip-diff-on-install --suppress-diff --wait -f k8s/helmfile.yaml`


Grafana: http://grafana.local
Mail: http://mail.local
