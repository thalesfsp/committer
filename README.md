# committer

It's a nimble and powerful tool that generates meaningful commit messages for you. Additionally, it allows you the developer to:
- Stage all files if none are staged
- Commit all staged files, if the generated commit message is acceptable, otherwise it can generate a new commit message, or allow you to enter your own
- Push the committed changes to the remote repository
- Tag the commit with a version number

## Install

### CLI

`curl -s https://raw.githubusercontent.com/thalesfsp/committer/main/resources/install.sh | sh`

Setting target destination:

`curl -s https://raw.githubusercontent.com/thalesfsp/committer/main/resources/install.sh | BIN_DIR=ABSOLUTE_DIR_PATH sh`

Setting version:

`curl -s https://raw.githubusercontent.com/thalesfsp/committer/main/resources/install.sh | VERSION=v{M.M.P} sh`

Example:

`curl -s https://raw.githubusercontent.com/thalesfsp/committer/main/resources/install.sh | BIN_DIR=/usr/local/bin VERSION=v1.3.17 sh`

### Programmatically

Install dependency:

`go get -u github.com/thalesfsp/committer`

## Usage

### CLI

`$ committer --help`

### Programmatically

See `*_test.go` files for examples.

### Documentation

Run `$ make doc` or check out [online](https://pkg.go.dev/github.com/thalesfsp/committer).

## Contributing

1. Fork
2. Clone
3. Create a branch
4. Make changes following the same standards as the project
5. Run `make ci`
6. Create a merge request

### Release flow

1. Update [CHANGELOG](CHANGELOG.md) accordingly.
2. Once changes from MR are merged.
3. Just tag. Don't need to create a release, it's automatically created by CI.

## Roadmap

Check out [CHANGELOG](CHANGELOG.md).
