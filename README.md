# unexport

Automatically unexport identifiers that are not used externally.

This tool will scan a Go repo, extract all names that are exported (start with uppercase letter), and produce a list of [gorename](https://github.com/golang/tools/blob/master/refactor/rename/rename.go) commands that rename the identifiers to lowercase.

## Usage

It is recommended that you start with a new GOPATH and only clone projects that you know are dependent on your target repo. This will save some time scanning repos.

This tool has two main functions:

### List mode

List mode, invoked with `unexport -l`, will print a list of identifiers in CSV format, with three columns:

1. Package name
2. Receiver (if applicable)
3. Name

In single-package repos, the package name will be constant, but it can change if you are operating on a package with `...` in the name.

The receiver is the parent struct of a field or func.

The name is obviously any name (constant, func, field, or variable), excluding interfaces (gorename seems to allow renaming interfaces in a way that will break your codebase).

Example:

```
unexport -l -p github.com/pilosa/pilosa/...
```

This prints all exported identifiers in CSV format to the command line.

### Run mode

Run mode, invoked with `unexport -r`, prints out the actual unexport commands. This may be piped to a shell if you actually want to run the commands, but likely you will want to modify the list of commands before running it, or perhaps run your test suite between each rename to ensure compatibility.

Run mode takes an additional argument, `-b`, to accept a blacklist of identifiers to filter out. The blacklist is in the same CSV format as produced by the list mode above.

Example:

```
unexport -r -b blacklist.csv -p github.com/pilosa/pilosa/...
```

### Workflow

A normal workflow with this tool is to first run list mode, redirecting the output to a file, then modify the file to exclude identifiers that you wish to keep in your public API, then pass that file in to `-b`. After producing the list of unexport commands, then you can either run that whole file as a bash script, or you can edit it to run your test suite between each command, or conditionally commit the changes to a version control system.
