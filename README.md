# transfer

Simple CLI tool to transfer files to transfer.sh or a self-hosted transfer.sh.

# Usage

`transfer file`

`transfer file --name test --downloads 1`

`transer folder`

`transfer folder --name test.zip --days 1`

# Config

A config file can be placed in `~/.transfer.json` or `~/.config/transfer/transfer.json` so CLI flags do not need to be provided each time.

Example config:

```json
{
    "url": "https://transfer.selfhosted.com",
    "user": "basic_auth_user",
    "pass": "basic_auth_pass"
}
```
