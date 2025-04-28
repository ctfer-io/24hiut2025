<div align="center">
    <h1>24h IUT 2025</h1>
    <a href="https://discord.com/channels/1333366010232705097/1333366010753056831"><img src="https://img.shields.io/badge/discord-24hiut25-5865F2?style=for-the-badge&logo=discord"></a>
    <!--<a href=""><img src="https://img.shields.io/github/license/ctfer-io/24hiut2025?style=for-the-badge" alt="License"></a>-->
    <a href="https://github.com/ctfer-io/24hiut2025/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-apache--2.0-green?style=for-the-badge"></a>
</div>

This repository contains the challenges and infrastructure elements for the [24h IUT 2025](https://24hinfo.iut.fr/).

> [!WARNING]
> This repository is a Work In Progress (CI, filesystem architecture, challenges). It is subject to major changes.
>
> In case of any non-retrocompatible or other breaking changes, [PandatiX](https://github.com/pandatix) will inform the relevant parties.

## Challenges

| Category | Name | Difficulty | Status | ChallMaker |
|---|---|---|---|---|
| Misc             | Memes 1/3               | Easy   | Ready    | FireFlan     |
| Misc             | Bottle Flip Challenge   | Easy   | Ready    | WildPasta    |
| Infra            | Kubrac                  | Medium | Ready    | PandatiX     |
| Web              | WordPressure            | Easy   | Ready    | WildPasta    |
| Web              | Sticky Match            | Medium | Incoming | WildPasta    |
| Web              | ?                       | Easy   | Incoming | BadZ_        |
| Web              | BlogCola 1/2            | Easy   | Incoming | walwal29     |
| Web              | BlogCola 2/2            | Medium | Incoming | walwal29     |
| Pentest          | Beverage Bazaar         | Easy   | Ready    | WildPasta    |
| Pentest          | Fatal Request           | Medium | Ready    | WildPasta    |
| Pentest          | L'illusionniste         | Easy   | Review   | fr4gments    |
| Reverse          | Reverse The Duck        | Easy   | Review   | Cya3gha      |
| Reverse          | Freizh Exam             | Easy   | Ready    | WildPasta    |
| Windows          | LAB AD 1                | Hard   | Incoming | KlemouLeZoZo |
| Windows          | LAB AD 2                | Insane | Incoming | KlemouLeZoZo |
| Crypto           | Vernam                  | Medium | Ready    | PandatiX     |
| Crypto           | Shamir                  | Medium | Ready    | PandatiX     |
| Crypto           | CBC-R encryption oracle | Hard   | Incoming | PandatiX     |
| Forensic         | Layers                  | Easy   | Ready    | PandatiX     |
| Forensic         | SodaStream              | Medium | Ready    | WildPasta    |
| Forensic         | The Vault Dweller       | Hard   | Review   | WildPasta    |
| Stega            | Le Planqué 2            | Easy   | Review   | PandatiX     |
| Stega            | Memes 2/3               | Medium | Ready    | FireFlan     |
| Stega            | Memes 3/3               | Hard   | Ready    | FireFlan     | 
| Threat Hunting   | ColAPT 1/4              | Medium | Ready    | hashp4       |
| Threat Hunting   | ColAPT 2/4              | Easy   | Ready    | hashp4       |
| Threat Hunting   | ColAPT 3/4              | Easy   | Ready    | hashp4       |
| Threat Hunting   | ColAPT 4/4              | Medium | Incoming | hashp4       |
| Pwn              | Ret2PopaCola            | Medium | Ready    | Souehda      |

Ideas:
- Poor Registry (Infra / Insane), NicoFgrx/PandatiX
- Kubroc (Infra / Hard), PandatiX
- Kubroken (Infra / Insane) PandatiX
- ? (Reverse/Pwn / ?) KlemouLeZozo
- ? (Prompt Injection / ?) BadZ_

### Team

- Admin
  - [PandatiX](https://github.com/pandatix)
  - [NicoFgrx](https://github.com/NicoFgrx)
  - [WildPasta](https://github.com/wildpasta)
- Ops
  - PandatiX
  - NicoFgrx
  - WildPasta
- ChallMaker
  - PandatiX (Infra)
  - WildPasta (Pentest)
  - KlemouLeZoZo (Windows)
  - d07pwn3d (OSINT)
  - hashp4 (Threat Hunting)
  - Cya3gha (Reverse)
  - Souehda (Pwn)
  - juju665937 (Multi Agent Systems)
  - FireFlan (Fun)
  - fr4gments (Pentest)

### Classification

Flag format: `24HIUT{...}`

Scoring:
- Score: **500** per challenge, **50** per side quest
- Decay: **15**
- Minimum: **50**

Difficulties:
- **Easy**: introduction level, everyone should be able to complete under 2 hours (with hints)
- **Medium**: require some knowledges, potentially acquired during the event with previous challenges
- **Hard**: require previous knowledges and creativity to solve
- **Insane**: require complex skills (might not be solved under 8 hours)

> [!NOTE]
> The 24h IUT 2025 targets BAC+1 to BAC+3 students, with mostly no previous experience in the field of cybersecurity.
> The event start at friday 2PM, then 8 hours of algorithmic challenges, 8 hours of web development, and 8 hours for the CTF (saturday 6AM-2PM).
>
> This must be considered in the difficulty rating by the ChallMaker. If any question, please contact Admins.

Status:
- **Incoming**
- **Review**
- **Ready**

### How to add a challenge ?

1. Clone the repository.
    ```bash
    git clone git@github.com:ctfer-io/24hiut2025.git && cd "$(basename "$_" .git)"
    ```

2. Create the directory for your challenge, on your own branch.
    ```bash
    git checkout -b <category>/<name>
    mkdir -p challenges/<category>/<name> && cd $_
    ```

3. Create your challenge configuration file.
    ```bash
    touch challenge.yaml
    ```
    You can add the schema to trigger auto-completion for ease of completion :wink:
    ```bash
    echo "# yaml-language-server: \$schema=file://$(cd ../../.. && pwd)/schema.json" > challenge.yaml
    ```

4. If your challenges require files to give players, create the `dist` directory.
    ```bash
    mkdir dist
    ```

5. If your challenge require infrastructure, create the `infra` directory.
    ```bash
    mkdir infra
    ```
    In case your challenge deploys challenges on demand, add the `Dockerfile`s such that we could rebuild the challenges.
    More on this topic will come later.

6. Submit your challenge through a [Pull Request](https://github.com/ctfer-io/24hiut2025/compare/main?template=challenge_pr.md).
