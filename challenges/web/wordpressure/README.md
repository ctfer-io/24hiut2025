Ce challenge utilise plusieurs images Docker à savoir :
- library/mysql:9.2.0 (internet)
- library/wordpress:php8.2-apache (internet)
- challenges/web/wordpressure-cli:v0.1.0 (custom)

Pour construire l'image Docker, suivre la procédure suivante:
```bash
REGISTRY=localhost:5000 # adapt 
cd challenges/web/wordpressure/infra
docker build . -f Dockerfile-cli -t $REGISTRY/challenges/web/wordpressure-cli:v0.1.0 
```

Le scénario est **custom** car nous ne pouvons pas utiliser le *ExposedMonopod* ou le *ExposedMultipod* proposés par Chall-Manager.
Pour compiler le scénario, suivre la procédure suivante:

```bash
cd challenges/web/wordpressure/infra/scenario
CGO_ENABLED go build -o main main.go
zip -r scenario.zip main Pulumi.yaml
```
