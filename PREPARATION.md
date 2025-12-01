- have kind installed
- have cloud-provider-kind installed: https://kind.sigs.k8s.io/docs/user/loadbalancer/
- have helm-diff installed: `helm plugin install https://github.com/databus23/helm-diff`
- have nginx-ingress installed: `kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml`
- have the following lines in /etc/hosts:
```
127.0.0.1	mail.local
127.0.0.1	grafana.local
```
- create a cluster: `kind create cluster --config cluster.kind`
- apply all: `helmfile apply --skip-diff-on-install --suppress-diff --wait -f k8s/helmfile.yaml`
- apply step: `helmfile apply --suppress-diff --wait -f k8s/helmfile-0-prerequisites.yaml`


Grafana: http://grafana.local
Mail: http://mail.local
