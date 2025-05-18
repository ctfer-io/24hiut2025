# Write-Up - Misc / Bottle Flip Challenge

On a une archive qui contient *beaucoup* de fichiers et un flag se cache dedans.

Pour voir dans quel fichier se trouve notre objectif, il faut faire un peu de tri.
Et pour cela, on se base sur les hash de fichiers qui sont uniques :

```bash
# on trouve le hash le plus fréquent
md5sum * | tail
---
663d6cba8e6f584b4af97a4e53b442ab  challenge_images/24HIUT{FreizhCola_Bottle_fd8bef23328a}.jpg
663d6cba8e6f584b4af97a4e53b442ab  challenge_images/24HIUT{FreizhCola_Bottle_fd24b3fbf0cd}.jpg
[SNIP]

# on l'exclu de notre recherche
md5sum * | grep -v 663d6cba8e6f584b4af97a4e53b442ab
---
922d59d4df9a7f36b29ec1df05269925  challenge_images/24HIUT{FreizhCola_Bottle_9989bd1dd6ea}.jpg
```

Un seul fichier est sorti, c'est celui qui nous intéresse.
On peut ensuite le `strings` ou `exiftool` au choix :

```bash
# en utilisant strings
strings challenge_images/24HIUT{FreizhCola_Bottle_9989bd1dd6ea}.jpg
---
JFIF
.Exif
Upside_Down_Bottle_Found!
[SNIP]

# en utilisant exiftool
exiftool challenge_images/24HIUT{FreizhCola_Bottle_9989bd1dd6ea}.jpg
---
ExifTool Version Number         : 13.10
File Name                       : 24HIUT{FreizhCola_Bottle_9989bd1dd6ea}.jpg
Directory                       : challenge_images
[SNIP]
Image Description               : Upside_Down_Bottle_Found!
```
