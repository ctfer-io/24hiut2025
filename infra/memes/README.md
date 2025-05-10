
```bash
cd infra/memes
docker build . -t registry.dev1.ctfer-io.lab/challenges/fun/memes:v0.1.0
docker push registry.dev1.ctfer-io.lab/challenges/fun/memes:v0.1.0

kubect create ns memes
kubectl create secret tls 24hiut2025-tls --cert=/etc/letsencrypt/archive/24hiut2025.ctfer.io/fullchain1.pem --key=/etc/letsencrypt/archive/24hiut2025.ctfer.io/privkey1.pem -n memes
kubectl apply -f deploy.yaml -n memes
```