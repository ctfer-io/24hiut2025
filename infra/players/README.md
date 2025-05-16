# Players

This submodule performs the auto-creation of users, teams, CTFd brackets (i.e. `interns` for non-competing participants, and `students` for competing participants) and captive portal users on the main firewall.

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
pulumi config set "forti-address" "10.17.30.254"
pulumi config set "forti-token" "api-token-key" --secret

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

On CTFd, the final step is to configure your CTFd to not accept further registrations, as all accounts should be managed per this Pulumi program.
On the firewall, the final step is to configure the interface to use the generated group to authenticate users.
