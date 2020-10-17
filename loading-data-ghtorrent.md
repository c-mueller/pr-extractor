# Loading Data from GHTorrent


## Step 0: Prerequisites

### Our Setup

- 32 GB of RAM
    - Less might work
- 6 Core CPU (AMD Ryzen R5 3600)
    - Lower specced CPUs also work, but with potentially lower performance
- 1TB of SSD storage for MongoDB (Formatted with XFS)
    - XFS is optional, however MongoDB recommends using it in conjuction with the WiredTiger storage engine.
    - Required
- 3TB of Archival storage to store the raw dumps
    - Not needed, if the dump files are deleted after restore
- A Third drive, to speed up the extraction of the `.tar.gz` archives.
    - We used the servers internal Operating System drive for this, since approx. only 50 GB of free space are needed. 
    - It can be considered optional, however it will drastically increase extraction speeds, if your archival storage is hard drive based.
- approx. 100 GB of SSD storage for the SQL database
- Fedora 32 Server Edition (64 Bit)
    - Other linux based systems definately work as long as they are Docker capable
- Docker
    - May be optional

### Running MongoDB

To run MongoDB we used the following Docker command:
```bash
docker run -d --name mongo --restart always -v /mongo:/data/db -p 27017:27017 mongo
```

Please replace `/mongo` with the location you want MongoDB to store the data at.

## Step 1: Downloading the MongoDB Dumps

Copy the links in [Files used](#files-used) into a textfile and download them using `wget -i`, alternatively you can extend the script used in Step 2 to download the current file, restore it, and then delete it. This can be done, if you do not have enough archival storage available.

## Step 2: Collect MongoDB Records

To extract and restore the data from the dumps, we have used the following Fish script:
```fish
for i in *.tar.gz
    echo $i
    pv $i | tar xvz --exclude="dump/github/pull_request_commits.bson" --exclude="dump/github/pull_request_commits.metadata.json" --exclude="dump/github/commits.bson" --exclude="dump/github/commits.metadata.json" -C /home/chris/dumps/
    echo $i
    mongorestore -h 127.0.0.1:27017 /home/chris/dumps/dump
    rm -rv /home/chris/dumps/dump
    mv -v $i dumped/
end
```

This script assumes, you are in the directory, in which all the downloaded dumps are located and that a child directory called `dumps/` has been created. in the following we will explain what this script exactly does:
```
pv $i | tar xvz <<EXCLUDES>> -C /home/chris/dumps/
```
extracts the Tar Arcive in to the directory after the `-C` flag, please adjust this one accordingly to fit your needs. The excludes, in the script, explicitly exclude large files that we do not need for our study. 

```
mongorestore -h 127.0.0.1:27017 /home/chris/dumps/dump
```
restores the extracted `bson` files into mongo db. Please adjust the directory accordingly to the one chosen before, but you have to make sure `/dump` is appended, otherwise, `mongorestore` will not restore the contents of the dump extracted.

If the restore process is executed on another machine, please also ensure to change the IP address accordingly.

```
rm -rv /home/chris/dumps/dump
```
removes the previously restored dump files. please make sure the directory matches the one used for `mongorestore`.

```
mv -v $i dumped/
```
moves the dump file in the `dumped/` directory, to make sure it is kept in case the process has to be repeated. 

Otherwise you can also replace this with `rm -v $i` to delete the file.

## Step 3: (Optional) Create a `type` index

To speed up the data extraction process we created an index for every document in the `events` collection used to identify the type of event. This helps speed up the search for elements we are interessted in (PullRequestEvents and PullRequestReviewCommentEvents). Creating an index is done by using the `mongo` command as follows:
After running `mongo 127.0.0.1:27017` enter the following commands:
```javascript=
use github
db.events.createIndex({type: 1}, {name: "type"})
```
this may take a while, the CLI can be closed, however make sure the the background process is not terminated by pressing `n` when the cli asks about that.

The progress of the index generation process can be checked by running the following command (assuming no connection i.e. restore is running at the same time):
```
docker logs mongo | tail
```

If you are not using docker the progress can be found in the logs of `mongod` for example by using `journalctl`.

## Step 4: Setting up PostgreSQL

To run a instance of PostgreSQL we used the following Docker command:
```
docker run --name pulls -v pulls_db:/var/lib/postgresql/data -p 5432:5432 -e POSTGRES_DB=pulls -e POSTGRES_PASSWORD=pulls -e POSTGRES_USER=pulls -d postgres
```

*Please Note*: Other databases like MySQL or SQLite will also work, however our query set was not tested against these DBMS for that reason we reccommend using PostgreSQL

## Step 4: Transform the data to SQL using the `pr-extractor` utility

First download and compile the `pr-exctractor` utiltiy from here: https://github.com/c-mueller/pr-extractor

The instructions, how to compile the utility can be found in the readme of the repository.

After compiling a config file has to be created in our case the file (`config.yml`) looked like this:
```yaml
# Defines the MongoDB url when using docker, only the IP / Hostname has to be changed here
mongo_url: mongodb://archie.l.krnl.eu:27017/?compressors=zlib&readPreference=primary&gssapiServiceName=mongodb&ssl=false
# Declares the used Database Driver for the GORM mapper. The utility supports 'sqlite3', 'mysql' and 'postgres'
db_driver: postgres
# Database URL: Please consult GORM docs to find out how this parameter has to look for each DBMS
# When using PostgreSQL in docker only IP / Hostname and Port may have to be changed
db_url: user=pulls dbname=pulls host=127.0.0.1  password=pulls port=5432 sslmode=disable
# True for verbose logs
verbose: true
# The number of workers, this generally does not have to be changed, however it must be included in the configuration file
worker_count: 4
# The number of elements in a worker queue, this generally does not have to be changed, however it must be included in the configuration file
worker_queue_length: 250
```

Once the config file is written to disk run:
```
./pr-extractor <Path to the config file>
```

Assuming you are in same directory in which the binary is located.

This command will take a while. In our case ~36 hours.


## Step 5: Optimizing PostgreSQL query Performance

## Appendix

### Files used

```
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-01-14.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-01-15.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-01-16.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-01-17.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-11.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-12.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-13.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-14.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-15.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-16.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-17.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-18.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-19.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-20.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-21.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-22.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-23.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-24.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-25.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-26.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-27.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-28.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-29.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-04-30.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-01.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-02.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-03.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-04.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-05.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-06.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-07.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-08.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-09.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-10.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-14.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-15.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-16.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-17.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-18.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-19.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-20.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-21.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-22.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-23.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-24.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-25.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-26.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-27.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-28.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-29.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-30.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-05-31.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-01.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-02.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-03.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-04.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-05.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-07.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-08.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-09.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-10.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-11.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-12.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-13.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-14.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-15.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-16.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-18.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-19.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-20.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-21.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-22.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-23.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-24.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-25.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-26.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-27.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-28.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-29.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2019-06-30.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-01-19.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-01-20.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-01-21.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-01-22.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-01-23.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-01-24.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-01-25.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-01-26.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-01-27.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-01-28.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-01-29.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-01-30.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-01-31.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-04.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-05.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-06.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-07.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-08.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-09.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-10.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-11.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-12.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-13.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-14.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-15.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-16.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-17.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-18.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-19.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-20.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-22.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-23.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-24.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-25.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-26.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-27.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-28.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-02-29.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-03-01.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-03-02.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-03-03.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-03-04.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-03-05.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-03-06.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-03-07.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-03-08.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-03-09.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-03-10.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-03-31.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-01.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-02.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-03.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-04.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-05.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-06.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-07.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-08.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-09.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-10.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-11.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-12.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-13.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-14.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-15.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-16.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-17.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-18.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-19.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-20.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-21.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-22.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-23.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-24.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-25.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-26.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-27.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-28.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-29.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-04-30.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-01.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-02.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-03.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-04.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-05.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-06.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-07.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-08.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-09.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-10.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-11.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-12.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-13.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-14.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-15.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-16.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-17.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-18.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-19.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-20.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-21.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-22.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-23.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-25.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-26.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-27.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-28.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-29.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-30.tar.gz
http://ghtorrent-downloads.ewi.tudelft.nl/mongo-daily/mongo-dump-2020-05-31.tar.gz
```
