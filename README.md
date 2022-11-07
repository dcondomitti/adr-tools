# ADR Tools

This is a collection of tools for working with Architecture Decision Records (ADRs).

## Features

### Index of ADRs

Given a directory of `decisions/`, the `rebuild-index` command will open a pull request to update the `README.md` file
with a table of contents of decision records.  This index will include records merged to main and records in open pull
requests labeled `proposal`.

## Usage

```bash
$ adr-tools help
NAME:
   adr-tools - A tool for working with Architecture Decision Records

USAGE:
   adr-tools [global options] command [command options] [arguments...]

COMMANDS:
   rebuild-index  Rebuilds the index of ADRs
   help, h        Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help (default: false)
```
