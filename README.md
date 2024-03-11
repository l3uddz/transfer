# transfer

Simple CLI tool that allow upload files and directories to transfer.sh. As an option it supports different hosts, credentials and download limits.

# Cli arguments

### Flags:
  -h      | --help            | Env vars            | Show context-sensitive help.       
----------|-------------------|---------------------|----------------------------------
  &nbsp;  | --version         | &nbsp;              | Print version information and quit 
  &nbsp;  | --update          | &nbsp;              | Check for an updated version       
  -u      | --url=transfer.sh | TRANSFER_URL        | Transfer.sh Service URL            
  &nbsp;  | --user=STRING     | TRANSFER_USER       | Transfer.sh Basic Auth Username    
  &nbsp;  | --pass=STRING     | TRANSFER_PASS       | Transfer.sh Basic Auth Password    
  -d      | --downloads=INT   | TRANSFER_DOWNLOADS  | Maximum amount of downloads        
  -D      | --days=INT        | TRANSFER_DAYS       | Maximum amount of days             
  -n      | --filename=STRING | &nbsp;              | Name of file when uploaded         
     

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

A config file can be placed in `~/.transfer.yml` or `~/.config/transfer/transfer.yml` so CLI flags do not need to be provided each time.

Example config:

```yaml
url: https://transfer.selfhosted.com
user: basic_auth_user
pass: basic_auth_pass
max_downloads: 1
max_days: 1
```
