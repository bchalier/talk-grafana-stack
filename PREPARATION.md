- have kind installed
- have cloud-provider-kind installed: https://kind.sigs.k8s.io/docs/user/loadbalancer/
- have nginx-ingress installed: `kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/main/deploy/static/provider/kind/deploy.yaml`
- have the following lines in /etc/hosts:
```
127.0.0.1	mail.localhost.local
127.0.0.1	grafana.localhost.local
```
- create a cluster: `kind create cluster --config cluster.kind`
