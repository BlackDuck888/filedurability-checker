# filedurability-checker
Storj V3 Network File Durability Checker

##Install

```
git clone https://github.com/stefanbenten/filedurability-checker
go install ./...
```

## Usage

```
filedurability-checker --file awesomedurable.file --apikey <YOURAPIKEY>
```

The following flags are supported:

```
--addr     -  Satellite Address, default: satellite.stefan-benten.de:7777
--apikey   -  API Key, no default
--enckey   -  Encryption Key, default: you'll never guess this
--bucket   -  Bucket Name, default: file-durability
--path     -  Path in the Bucket, no default
--file     -  File to upload (specify from your local folder)
--interval -  Interval between the checks in seconds, default: 300
--listen   -  Webserver Listen Address, default: :8080
```