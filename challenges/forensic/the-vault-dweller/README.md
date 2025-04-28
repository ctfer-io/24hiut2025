# THE VAULT DWELLER

Dans ce challenge forensic, on doit analyser une capture RAM Windows 11 pour extraire une base de mots de passe chiffrée utilisée par un gestionnaire personnalisé (pcola-vault-manager).

Notions abordées :

- Lister et identifier les processus en mémoire avec Volatility3
- Dumper un fichier .passdb en gardant à l'esprit la notion de handle ouvert
- Analyser statiquement un binaire pour comprendre une logique de chiffrement custom (XOR + Base64)
- Extraire un mot de passe en clair depuis un processus Notepad.exe en exploitant la persistance des onglets ouverts (TabState)
- Déchiffrer manuellement la base de données
