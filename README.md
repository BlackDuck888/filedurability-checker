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

_Note: In this version it is not checking if the bucket is existing. Please create it beforehand._

### The following flags are supported:

#### Required:
```

--apikey   -  API Key, no default
--file     -  File to upload (specify from your local folder)
```
#### Optional:
```
--addr     -  Satellite Address, default: satellite.stefan-benten.de:7777 
--enckey   -  Encryption Key, default: you'll never guess this
--bucket   -  Bucket Name, default: file-durability
--path     -  Path in the Bucket, no default
--interval -  Interval between the checks in seconds, default: 300
--listen   -  Webserver Listen Address, default: :8080
```