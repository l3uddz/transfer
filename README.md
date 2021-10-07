# transfer

Simple CLI tool that allow upload files and directories to transfer.sh. As an option it supports different hosts, credentials and download limits.

# Cli arguments

### Flags:
  -h  | --help              |   Show context-sensitive help.
  ----|---------------------|-----------------------------------
      |  --version          |   Print version information and quit
      |  --update           |  Check for an updated version
  -u  | --url=transfer.sh |   Transfer.sh Service URL
      |  --user=STRING      |  Transfer.sh Basic Auth Username
      |  --pass=STRING      |  Transfer.sh Basic Auth Password
  -d  | --downloads=INT     |   Maximum amount of downloads
  -D  | --days=INT          |   Maximum amount of days
  -n  | --filename=STRING   |   Name of file when uploaded



# Usage

`transfer img.png`

>Download URL: https://transfer.sh/123/img.png
>
>--
>
>Delete URL: https://transfer.sh/123/img.png/123321123

`transfer img.png --name test --downloads 1`

>Download URL: https://transfer.sh/123/test
>
>--
>
>Delete URL: https://transfer.sh/123/test/123321123

`transer directory`

>Download URL: https://transfer.sh/123/directory.zip
>
>--
>
>Delete URL: https://transfer.sh/123/directory.zip/123321123

`transfer directory --name test.zip --days 1`

>Download URL: https://transfer.sh/123/test.zip
>
>--
>
>Delete URL: https://transfer.sh/123/test.zip/123321123

# Config

A config file can be placed in `~/.transfer.json` or `~/.config/transfer/transfer.json` so CLI flags do not need to be provided each time.

Example config:

```json
{
    "url": "transfer.selfhosted.com",
    "user": "basic_auth_user",
    "pass": "basic_auth_pass"
}
```
