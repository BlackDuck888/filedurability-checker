# Storj V3 Network File Durability Checker

## What can it do

This tool uploads a file of your choice and then downloads it every (default) 5 Minutes to check if the file is still
available. After download it compares it with the original file to ensure its correct.

It serves a website (default :8080) with the statistics about the test, as well as the errors if it failed on a specific
 check.
 
 The idea of this tool is to be ran with multiple file sizes on different locations, to ensure network stability and
  durability.

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

_More features to come..._
