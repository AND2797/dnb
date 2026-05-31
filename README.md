# dnb: daily notebook

a small CLI for keeping daily notes. it creates a text file in the
`YYYYMMDD.txt` format under a `YYYY/MM/` directory structure, so ideas, todos,
etc. stay organized by day. when you open a new day it rolls over the previous
day's contents (under a fresh date header) to keep a continuous history.

## usage

```
dnb list                 # list configured notebooks
dnb open <notebook>      # open today's file for a notebook (in $EDITOR, default vim)
```

## config

create `~/.dnbconf/config.yaml` (see `example-config.yaml`):

```yaml
notebook_root: ~/dnb_notebooks
notebooks:
  - daily_notebook
  - personal_notebook
```

## build

```
go build ./cmd/dnb
go test ./...
```
