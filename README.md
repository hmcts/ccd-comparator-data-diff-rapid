[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![test](https://github.com/hmcts/ccd-comparator-data-diff-rapid/workflows/ccd-comparator-data-diff-rapid/badge.svg "test")

# ccd-comparator-data-diff-rapid

## Prerequisites
- [Golang 1.21.0](https://go.dev/doc/install)
- PostgreSQL Database connection. See below for required variables


## Required Environment Variables
Replace the values to match your local setup
```shell
export DATABASE_HOST=localhost
export DATABASE_PORT=5050
export DATABASE_USERNAME=ccd
export DATABASE_PASSWORD=ccd
```

## Using the executable
To run the exec, use the following command:
```shell
go build -ldflags "-s -w"
```
```shell

go run .
```
Capture and display output on the console while removing escape characters and saving logs to 'output.log'
```shell

go run . | tee >(sed 's/\x1b\[[0-9;]*[a-zA-Z]//g' > output.log)
```


### Available CLI Arguments
Use as follows: `go run . -help`

| Argument                            | Description                                         |
|-------------------------------------|-----------------------------------------------------|
| `-configFile`                       | Configuration file (default "./config")             |
| `-sourceFile`                       | File contains existing caseTypes with jurisdictions |
| `-mem-profile file`                 | Write memory profile to file                        |
| `-cpu-profile`                      | Write cpu profile to file                           |


### Case Filtering

* **Jurisdiction**: The jurisdiction for filtering cases by. This should be passed in as a string representing the
  jurisdiction.
  If this parameter is passed, it will ignore scanCaseReferences and look for:
  *  **Period Start Date**: The start time of the time range for filtering cases by event details. 
  This should be passed in as a LocalDateTime object in the format "yyyy-MM-dd'T'HH:mm:ss".
  * **Period End Date**: The end time of the time range for filtering cases by event details. 
  This should be passed in as a LocalDateTime object in the format "yyyy-MM-dd'T'HH:mm:ss".
  * **Case Type**: The case type for filtering cases by. This should be passed in as a string representing the case type.

* **Include Empty Change**:  A boolean flag indicating whether to include empty change lines in the Excel report, 
regardless of whether the change violates the rule or not. This should be passed in as true or false. 
If set to true, the Excel report will include all change lines, which can be useful for narrow filters or when using with case reference search, 
as it may produce a large number of rows in the Excel report. 

