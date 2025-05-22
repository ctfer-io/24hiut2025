## Launch locally

Requirements:
- [docker](https://docs.docker.com/engine/install/)
- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [pulumi](https://www.pulumi.com/docs/iac/download-install/)
- [jq](https://jqlang.org/download/)

```bash
# Construct docker image
docker network create kind || true
docker run -d --network kind --name registry -p 5000:5000 registry:2
(cd webserver && docker build -t localhost:5000/infra/kubrac:v0.1.0 .)
docker push localhost:5000/infra/kubrac:v0.1.0
docker pull busybox && docker tag busybox:latest localhost:5000/busybox:latest && docker push localhost:5000/busybox:latest

# Create the cluster with ingress controller
kind create cluster --config=kind-config.yaml
kubectl apply -f https://kind.sigs.k8s.io/examples/ingress/deploy-ingress-nginx.yaml

# (Optional) Add DNS entry to ease challenge access
sudo echo "$(ip -4 addr show docker0 | grep -oP '(?<=inet\s)\d+(\.\d+){3}') a0b1c2d3.24hiut2025.ctfer.io" >> /etc/hosts

# Configure stack and deploy
cd deploy
export PULUMI_CONFIG_PASSPHRASE=""
pulumi login --local # this may be skipped if you already use pulumi
pulumi stack init kubrac
pulumi config set identity a0b1c2d3
pulumi config set --path "additional.registry" "localhost:5000"
pulumi up --yes

# You can then reach the challenge !
eval "$(pulumi stack output -j | jq -r ".connection_info")"
```

## Wipe it properly

```bash
pulumi dn --yes
pulumi stack rm --yes
cd ..
kind delete cluster
docker stop registry && docker rm registry
```
