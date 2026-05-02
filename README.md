# Backup GitHub

A simple CLI tool to back up all of your GitHub repositories as local mirror clones.

## Features

- Backs up all repositories (public, private, and forks) for your authenticated GitHub account
- Uses `git clone --mirror` to create complete mirror clones including all branches and refs
- Stores OAuth tokens securely in your system keyring; no config files with plain-text credentials
- Handles GitHub API pagination automatically

## Prerequisites

- Go 1.22 or later
- Git installed and available in `PATH`
- A GitHub account

## Installation

### Homebrew (macOS)

```sh
brew tap ChrisWiegman/backup-github
brew install backup-github
```

### Go install

```sh
go install github.com/ChrisWiegman/backup-github/cmd/backup-github@latest
```

### Download a release

Pre-built binaries for macOS are available on the [Releases](https://github.com/ChrisWiegman/backup-github/releases) page.

## Usage

Run the command from the directory where you want your backups to be saved:

```sh
backup-github
```

On first run you will be prompted to authenticate with GitHub via OAuth. The token is saved to your system keyring so subsequent runs do not require re-authentication.

Repositories are cloned into a `backups/` subdirectory of your current working directory:

```
./backups/
├── my-repo/
├── another-repo/
└── forked-repo/
```

Each directory is a bare mirror clone. Re-running `backup-github` will clone any new repositories and update existing ones via `git remote update`.

### Version information

```sh
backup-github --version
backup-github -v
```

## GitHub OAuth permissions

The app requests the following GitHub scopes during authentication:

| Scope | Purpose |
|-------|---------|
| `repo` | Access private repositories |
| `read:org` | Read organization membership |
| `gist` | Access gists |

## Development

### Build

```sh
make install
```

This compiles the binary with the current git tag as the version and installs it to `$GOPATH/bin`.

### Test

```sh
make test
```

### Lint

```sh
make lint
```

Requires [golangci-lint](https://golangci-lint.run/). The linter is installed automatically if not found.

### Update dependencies

```sh
make update
```

### Changelog

This project uses [Changie](https://changie.dev/) for changelog management.

Add a new changelog entry:

```sh
make change
```

Cut a release changelog:

```sh
make changelog
```

## License

MIT — see [LICENSE.txt](LICENSE.txt).
