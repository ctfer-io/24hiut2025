# Write-Up - Stega / Le planqué 2

Un article scientifique sur le Coca-Cola ? Curieux. Mais il semble cacher bien plus qu’une simple étude clinique. Explorez-le en profondeur pour y trouver un fichier caché. Déverrouillez ce dernier pour révéler le flag.

## Analyse du PDF avec binwalk

On commence par scanner le fichier PDF à la recherche de fichiers dissimulés à l’intérieur.

```bash
binwalk -e Systematic_review_Coca-Cola_can_effectively_dissolve_gastric_phytobezoars_as_a_first-line_treatment.pdf
cd _Systematic_review_Coca-Cola_can_effectively_dissolve_gastric_phytobezoars_as_a_first-line_treatment.pdf.extracted
```

## Extraction du hash avec zip2john

```bash
zip2john 4A4132.zip > hash.out
```

## Brute-force avec John the Ripper

```bash
john --wordlist=/usr/share/wordlists/rockyou.txt hash.out
---
Using default input encoding: UTF-8
Loaded 1 password hash (PKZIP [32/64])
Will run 4 OpenMP threads
Press 'q' or Ctrl-C to abort, almost any other key for status
arviegrace       (4A4132.zip/flag.txt)     
```

## Récupérer le contenu du flag

```bash
unzip -P arviegrace 4A4132.zip
cat flag.txt
```
