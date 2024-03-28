# mrca

Find the most recent common ancestor (MRCA) of a set of tips in a phylogenetic tree

## Use

The input tree (`-t`) must be rooted and in newick format, with internal tips labelled. Tip names are matched by providing a regular expression to `-r`:

```
> mrca -t rooted.newick -r "in|out0"
Node2
```

The MRCA node's label is written to `stdout`.

#### Help:

```
./mrca -h
Usage:
   [flags]

Flags:
  -h, --help           help for this command
  -r, --regex string   regex of tip names to parse
  -t, --tree string    tree file to read (in Newick format
```

## Installation

First, [install go](https://go.dev/dl/),

then:

```
go install github.com/bjeight/mrca@latest
```

or

```
git clone https://github.com/bjeight/mrca.git
cd mrca
go build -o mrca
```