# Instructions pour les challenges "LAB"

Pour ce challenge, vous aurez accès à une VM Linux raccordée à 2 réseaux :
- le premier vous permet de vous connecter à distance à la VM ;
- le deuxième est raccordé à un réseau en `25.0.x.0/24` qui contient plusieurs autres VMs.

Pour vous y connecter, vous êtes libres de choisir entre les 2 méthodes ci-après.

## Connexion SSH

Depuis CTFd, après avoir déployé votre instance, des informations de connexion vous seront fourni sur la pop-up du challenge :
1. Copier/coller la commande SSH `ssh -l user 10.17.XX.XX` ;
2. Entrer le mot de passe affiché.

Vous pouvez maintenant travailler depuis la VM.

## Connexion VPN

Vous avez la possibilité de vous connecter via VPN pour plus facilement utiliser les outils de votre machine :
1. Se connecter via SSH (cf étapes plus haut) et récupérer le fichier `client.ovpn` ;
2. Importer le fichier (cf https://openvpn.net/connect-docs/import-profile.html).

Si vous souhaitez vous connecter à plusieurs, générez un autre `client.ovpn` en utilisant le script `/home/user/openvpn.sh`.
