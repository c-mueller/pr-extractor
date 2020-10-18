# Loading Data from GHArchive

The `pr-extractor` utility also supports event import from json files passed via `stdin`. This allows the import of the data from GHArcive or other event sources.

NOTE: This guide is just a small complementary guide to the [GHTorrent Guide](loading-data-ghtorrent.md) please refer to it for the Setup of PostgreSQL and the creation of the configuration file. Unlike GHTorrent MongoDB is not needed for this approach.

Loading data can be done using the following Fish script:

```fiah
for i in (bash -c "echo https://data.gharchive.org/2019-{01..12}-{01..31}-{0..23}.json.gz | sed 's| |\n|g'")
	echo Processing $i
	curl -f $i | gunzip | ./pr-extractor json -c config.yml || true
end
```

Explanation:

- `bash -c "echo https://data.gharchive.org/2019-{01..12}-{01..31}-{0..23}.json.gz | sed 's| |\n|g'"` Generates a list of all files available for a year, please adjust the numbers to fit your needs, in our case this process was  calĺed twice_
	- `bash -c "echo https://data.gharchive.org/2019-{01..12}-{01..31}-{0..23}.json.gz | sed 's| |\n|g'"`
	- `bash -c "echo https://data.gharchive.org/2020-{01..05}-{01..31}-{0..23}.json.gz | sed 's| |\n|g'"`
- `curl -f $i | gunzip | ./pr-extractor json -c config.yml || true` downloads the file, unpacks it and loads it into the database using the specified config
	- the use of the `-f` flag is important to fail the file does not exist. e.g. 31.02.2019
	- `|| true` prevents failure of the loop if the file does not exist, may be unnecessary.