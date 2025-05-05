# Players

This submodule performs the auto-creation of users, teams, and CTFd brackets (i.e. `interns` for non-competing participants, and `students` for competing participants).

## How to use

It consumes a JSON file named `players.json` structured as follows.
Notice:
- `.teams[].players.[].role` can be:
   - `admin` for administration accounts;
   - `intern` for players that are not competing, i.e. university interns, professors, graduates;
   - `companion` for guests coming with their team, but not participating to the event actively;
   - `player` for students who compete through the event.
- such file is under GDPR control, thus should **never** be commited nor transmitted to unauthorized parties. Rule of the thumb: we don't need it after the event, **destroy it ASAP**.

```json
{
    "teams": [
        {
            "name": "CTFer.io",
            "affiliation": "CTFer.io",
            "players": [
                {
                    "name": "PandatiX",
                    "email": "lucas.tesson@protonmail.com",
                    "role": "admin"
                }, {
                    "name": "NicoFgrx",
                    "email": "nicolas.faugeroux@protonmail.com",
                    "role": "admin"
                }, {
                    "name": "WildPasta",
                    "email": "richard.chauve@protonmail.com",
                    "role": "admin"
                }
            ]
        }, { ... }
    ]
}
```

You can run using the following. It will use a Service Account to populate your event.

```bash
# Create and configure the stack
export PULUMI_CONFIG_PASSPHRASE="please put a real password here"
pulumi stack init authn
pulumi config set "url" "https://24hiut.ctfer.io"
pulumi config set "username" "service-account-username"
pulumi config set "password" "service-account-password"

# Deploy it
pulumi up -y

# Export data to distribute credentials
pulumi stack output players --show-secrets | jq

# Destroy
pulumi dn -y
pulumi stack rm authn
```

> [!WARNING]
> Through the event players are expected to change their credentials.
> Please do not re-run the program once deployed, embrace the drift.

Final step is to configure your CTFd to not accept further registrations, as all accounts should be managed per this Pulumi program.

## Pipe into Fortinet 802.1x

The following command should generate the Fortinet 802.1x configuration script for the previously created users.
It **requires** a group `portail_captif` to be created before.

```bash
pulumi stack output players --show-secrets | jq -r '
  def escape_pw: gsub("\\\\"; "\\\\") | gsub("\""; "\\\"") | gsub("`"; "\\`");
  reduce .[] as $u ("config user local\n";
    . + "edit \"\($u.name)\"\nset type password\nset passwd \($u.password | escape_pw)\nnext\n"
  )
  + "end\n\nconfig user group\nedit portail_captif\nset member " 
  + (map(.name) | join(" ")) + "\nend"
'
```
