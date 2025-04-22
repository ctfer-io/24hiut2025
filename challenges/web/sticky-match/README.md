# Flask ReDoS Demo

Ce challenge est une application Flask qui démontre une vulnérabilité ReDoS (Regular Expression Denial of Service). 

## Structure du projet

```bash
flask-app
├── monitor.py       # Démarre app.py et supervise son exécution
├── app. py          # Présente l'application Flask aux joueurs
├── requirements.txt
├── static
│ ├── css
│ │ └── styles.css
│ └── js
│ └── scripts.js
├── templates
│ ├── flag.html      # Template pour afficher le flag
│ ├── home. html     # Template pour la page d'accueil
│ └── register.html  # Template pour la page d'enregistrement
└── README.md
```

## Instructions d'installation

1. **Clonez le dépôt** :

```bash
git clone https://github.com/ctfer-io/24hiut2025.git
cd 24hiut2025/challenges/web/sticky-match
```

2. **Exécuter l'application standalone** :

```bash
# prérequis : pip3 et python3
cd flask-app
pip3 install -r requirements.txt
python monitor.py --port 8080
```

3. **Exécuter l'application dans un Docker** :
```bash
# prérequis : Docker
docker compose -f ./docker-compose.yml up -d
```
