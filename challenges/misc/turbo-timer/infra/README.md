# SETUP

Install prerequisites:

```bash
python3 -m venv .venv
source .venv/bin/activate
pip3 install -r requirements.txt
```

Obfuscate the challenge according to your OS:

```bash
pyarmor gen --platform windows.x86_64 race.py
pyarmor gen --platform linux.x86_64 race.py
```

Don't forget to import the `assets` folder into the `dist` folder.
You should be able to run the game now.

```bash
python3 race.py
```
