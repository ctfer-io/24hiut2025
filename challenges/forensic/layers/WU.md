# Layers

Ce challenge est assez simple, et fait echo à une bonne pratique : Docker ne doit pas être utilisé, dès que possible, pour travailler avec des informations sensibles (clés SSH, GPG, authentification, construction de contenus). Si cela est impossible, il faut alors utiliser des secrets... Toutefois, dans la réalité, beaucoup d'entreprises ne prennent pas le temps de se poser ces questions, car il faut livrer pour rester concurrentiel, alors la sécurité passé à la trappe.

Heureusement, dans notre cas, nous voulons justement accéder au système d'authentification de PopaCola, et vu qu'ils ont voulu nous doubler, ils ont pris quelques raccourcis scénaristiques.

## Reconnaissance

Nous sommes face à un fichier, et nous déterminons sa nature.

```bash
$ file authenticator.tar 
authenticator.tar: POSIX tar archive
```

On peut untar le contenu pour fouiner un peu dedans et comprendre ce que cela contient. Note : ne pas faire cela si on ne sait pas d'où l'archive provient, et ce qu'elle devrait contenir. Cela aurait bien pu être une tar bomb :wink:

```bash
$ mkdir layers
$ tar -xf authenticator.tar -C layers
$ ls -al layers
$ ls -al layers
total 28
drwxrwxr-x 3 pandatix pandatix 4096 Apr 22 08:24 .
drwxrwxr-x 3 pandatix pandatix 4096 Apr 22 08:23 ..
drwxr-xr-x 3 pandatix pandatix 4096 Apr 22 08:13 blobs
-rw-r--r-- 1 pandatix pandatix  364 Jan  1  1970 index.json
-rw-r--r-- 1 pandatix pandatix 3616 Jan  1  1970 manifest.json
-rw-r--r-- 1 pandatix pandatix   31 Jan  1  1970 oci-layout
-rw-r--r-- 1 pandatix pandatix   98 Jan  1  1970 repositories
```

Normalement, on comprend rapidement que nous sommes face à une image Docker, et quelques recherches sur internet peuvent nous y aider.
D'après le nom du challenge et son énoncé, il semblerait que quelque chose se cache dans les _layers_ de l'image.
Plutôt que de fouiller manuellement, on va utiliser l'outil [`dive`](https://github.com/wagoodman/dive).

```bash
$ dive --source docker-archive authenticator.tar
```

On pourrait se balader longtemps, mais en descendant simplement jusqu'au dernier _layer_, on voit ̀ce qui suit.
```
Command:
RUN |1 KEY=ya-can-trust-me-broooo /bin/sh -c go build -o /go/bin/main -ldflags="-X 'main.Key=${KEY}'" main.go # buildkit
```

On va alors lancer l'image docker en question et donner la clé que nous vennons d'extraire.
Attention: ne pas oublier `-it` sinon nous n'avons pas de terminal interactif (TTY). On le comprend vite puisque ça plante sans cela...

```bash
$ docker load -i authenticator.tar
$ docker run -it authenticator:latest
```

En entrant la clé retourvée, on obtient `24HIUT{at least use multistep dockerfiles...}`.

Une solution alternative est d'extraire l'entrypoint `/go/bin/main`, de faire un coup de `strings`. Jugés plus complexes qu'un simple "coup de dive", ils sont acceptés et ne sont pas des _unexpected_.

## Recommandation

Dans ce cas précis, une méthode simple pour contourner ce problème d'inspection : faire une _multistep_ dans lequel on construit d'abord le binaire, puis une image minifiée (e.g. `scratch`) dans lequel on fait tourner le binaire, empêche une simple inspection des layers de nous donner le mot de passe. Toutefois, un peu de Reverse-Engineering et le tour est joué. La meilleure solution reste de ne pas chercher à construire un nouveau système d'authentification, mais utiliser des sytèmes bien rodés : Authenticator, Vault, ...
