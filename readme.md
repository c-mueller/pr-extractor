# `pr-extractor`

This small utility programm is used to transform the `events` collection from GHTorrents MongoDB dump files into a relational data model, that can be used for further analysis using SQL. The usage of this tool is described [here](http://google.com)
<!-- TODO: Insert URL to data generation manual -->

## Building `pr-exctractor`

To build the tool you must have the following tools installed:
-  Go Version 1.14 or later
-  gcc (Optional)
    -  Mandatory if you want to use SQLite as the relational database

After cloning the repository, the tool can be built using:

```
go build -v
```

