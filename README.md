# OS mirror data dive

## SLE12 Dive
### AIM
How many packages were updated in SLE12 (12 ~ 12-SP3).

### Data Source
There is not open mirror for SLE, so I try to get the full iso from download.suse.com.
The Archive.gz file is used in that ISOs.
These files are trunked and saved under `./sle/data`.

### Output

The output is saved under `./sle/output`.

## CentOS 7 Dive
### AIM
How many packages were updated in CentOS 7 (7.0 ~ 7.x).

### Data Source
From vault.centos.org, the primary meta data for example:
```
http://vault.centos.org/7.5.1804/os/Source/repodata/9f55a91fdf9499ec6ef222786316bd2be8372c9e999c807e386a52743e33b3ca-primary.xml.gz
```

The data are downloaded and saved under `./centos/data` .

### Output

The output is saved under `./centos/output`.
