## Write-Up - Stega / Memes 2/2

Explications techniques [ici](https://cyberhacktics.com/hiding-information-by-changing-an-images-height)

Tout d'abord, trouver l'image de la tombe Popa Cola : `resources/templates/images/3nx72a.jpg`

![Tombe Popa Cola](./wu/3nx72a.jpg)

L'ouvrir dans Cyberchef, la convertir en hexadécimal :

![cyberchef_1](./wu/cyberchef_input.png)

Passer l'hexadécimal en input, reconvertir en image et localiser la séquence `ff c0`, la taille est encodée 3 bytes plus loin:

![cyberchef_2](./wu/cyberchef_ouput.png)

Augmenter la taille à la main puis scroller sur l'image (oupsi la typo):

![cyberchef_3](./wu/cyberchef_ouput_final.png)

FLAG : `24HIUT{3nTeRR0nS_P0p4_C0L4}`
