# Shamir Write-Up

Dans ce challenge, nous devrions comprendre assez vite qu'un chat nommé "Shamir" ne le porte pas par hasard. Cela devrait rapidement axer les recherches autour du Shamir Secret Sharing (SSS) : une méthode pour séparer un secret en un nombre donné de parts, dont il faut en rassembler un sous-ensemble pour obtenir le secret.

Bien que peu fréquent, le SSS est en réalité une approche tant organisationelle que technique pour résoudre des problématiques de sécurité (e.g. il faut corrompre 3 administrateurs pour pouvoir passer Domain Admin).

## Approche brute-force

Si nous saurions nous débrouiller avec 4 clés (cf. énoncé), alors il semblerait que le seuil est de 4. Avec cette information, nous avons 2 approches :
1. On recherche les bits manquants à la 4ème clé, en sachant que le secret commence par `24HIUT` (Known Plaintext Attack / KPA) ;
2. On sait que le secret commence par `24HIUT`, pas besoin de chercher ce segment.

Dans une approche brute-force, il nous faudra retrouver 23 bytes. Cela représente 4951760157141521099596496896 combinaisons... Autant dire que c'est impossible, même avec l'information que la chaîne débute par `24HIUT`, car les bits manquants affectent en réalité le début du secret.

## Approche partielle

Puisqu'on connaît le début du secret, nous n'avons pas vraiment besoin de le chercher. Nous pouvons ainsi prendre n'importe quelle chaîne hexa random de 23 caractères, l'ajouter à la fin de la clé partielle, et corriger le début.

Par exemple, si on ajoute 23 `0`, on obtient un clé complète (bien qu'inexacte). En combinant les 4 clés, par exemple via https://iancoleman.io/shamir/, on obtient le secret `셕뙑圝䯣{Pas besoin de tout avoir, seulement un minimum}`.

En corrigeant le début, on obtient ainsi le flag `24HIUT{Pas besoin de tout avoir, seulement un minimum}`.

La preuve de cette approche est dispo dans `wu/index.js`.
