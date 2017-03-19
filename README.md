# gsr

Run git status recursively.

## Usage

By default, display only repositories with differences.

```bash
$ gsr path/to/dir
path/to/dir/repo1
path/to/dir/repo3
```

Show status with `--status` option.

```bash
$ gsr --status path/to/dir
path/to/dir/repo1
## master...origin/master
 M .gitignore

path/to/dir/repo3
## dev
 M cmd/root.go
```

If you use [motemen/ghq](https://github.com/motemen/ghq) , you can omit arguments.

```bash
$ gsr
/Users/yukimemi/.ghq/src/github.com/yukimemi/gsr
/Users/yukimemi/.ghq/src/github.com/yukimemi/core
```


## Install

To install, use `go get`:

```bash
$ go get github.com/yukimemi/gsr
```

## Contribution

1. Fork ([https://github.com/yukimemi/gsr/fork](https://github.com/yukimemi/gsr/fork))
1. Create a feature branch
1. Commit your changes
1. Rebase your local changes against the master branch
1. Run test suite with the `go test ./...` command and confirm that it passes
1. Run `gofmt -s`
1. Create a new Pull Request

## Author

[yukimemi](https://github.com/yukimemi)

