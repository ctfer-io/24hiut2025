# Write-Up - Stega / Mr Vernam

L'objectif de ce challenge est de comprendre que les renderers (ici pour PDF) ne sont fait que pour montrer ce qui a été décidé. Toutefois, du point de vue mémoire, ces formats permettent souvent d'utiliser des canaux auxiliaires et cachés.
Il fait echo à un [challenge similaire de l'édition 2023](https://github.com/pandatix/24hiut-2023-cyber/tree/main/stega/le-planque).

Après avoir téléchargé le fichier, on peut confirmer que c'est un PDF avec n'importe quel viewer. Le papier en lui-même n'a aucun rapport, ce qu'on devrait comprendre rapidement.
Toutefois, on peut chercher du contenu invisible via `strings Systematic\ review_\ Coca‐Cola\ can\ effectively\ dissolve\ gastric\ phytobezoars\ as\ a\ first‐line\ treatment.pdf | grep 24HIUT`, mais cela n'aboutira pas.

Pour continuer de fouiller dans ce PDF, nous pouvons aller chercher du côté des metadata, soit avec des outils en ligne (e.g. https://www.metadata2go.com), soit en local.

```bash
$ pdfinfo -custom Systematic\ review_\ Coca‐Cola\ can\ effectively\ dissolve\ gastric\ phytobezoars\ as\ a\ first‐line\ treatment.pdf
Author:          
CreationDate:    Thu Apr 24 22:35:33 2025 CEST
Creator:         Mozilla Firefox 128.9.0
Flag:            24HIUT{Le sachiez vous : PDF ne veut pas dire Pas De Flag}
Keywords:        
ModDate:         
Producer:        cairo 1.18.0 (https://cairographics.org)
Subject:         
Title:           
```

Il y a une metadata custom (`Flag`) qui contient le flag `24HIUT{Le sachiez vous : PDF ne veut pas dire Pas De Flag}`.
