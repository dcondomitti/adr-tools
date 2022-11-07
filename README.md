# ADR Tools

This is a collection of tools for working with Architecture Decision Records (ADRs).

## Features

### Index of ADRs

Given a directory of `decisions/`, the `rebuild-index` command will open a pull request to update the `README.md` file
with a table of contents of decision records.  This index will include records merged to main and records in open pull
requests labeled `proposal`.

## Usage

```
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

## Docker

```
$ docker run -ti bmorton/adr-tools adr-tools --help
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

## GitHub Actions

### Rebuild README.md directly in main

This GitHub Action will rebuild the `README.md` file in the main branch of your repository whenever a pull request is
merged or labeled with `proposal`.

If you'd like to rebuild with a randomly created branch and a pull request, remove the `--target-branch` and
`--pull-request` arguments from the example.

```yaml
name: Rebuild index

on:
  pull_request:
    types: [ labeled, unlabeled, closed ]

jobs:
  rebuild-index:
    if: ${{ github.event.label.name == 'proposal' || github.event.pull_request.merged == true }}
    runs-on: ubuntu-latest
    container:
      image: bmorton/adr-tools:latest
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        GITHUB_REPOSITORY: bmorton/architecture-example
    steps:
      - name: Rebuild index
        run: adr-tools rebuild-index --target-branch="main" --pull-request=false
```
