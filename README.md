# gitmirror

## Overview

`gitmirror` is a command-line interface (CLI) application designed to synchronize git repositories. It provides functionalities for discovering repositories, mirroring them, and keeping them in sync between the two platforms.

Modes:
- `sync`: Synchronize changes between origin and dsetination.
- `config`: Use a configuration file to specify repositories and settings.
- `discover`: Search for repositories in a provider and sync them to a destination.
> Only GitHub

Supported providers:
- GitHub
- GitLab

## Installation

```console
wget https://github.com/Desvelao/gitmirror/releases/download/v0.0.1-alpha1/gitmirror-0.0.1-alpha1-linux-amd64.tar.gz \
&& tar -xzf gitmirror-0.0.1-alpha1-linux-amd64.tar.gz \
&& sudo mv gitmirror /usr/local/bin/
```

See [Releases](https://github.com/Desvelao/gitmirror/releases) for the latest version and installation instructions for other platforms.

Check installation:
```console
gitmirror --help
```

## Usage

Synchronize repositories:

- Origin and destination specified as command-line arguments:
```console
gitmirror sync git:github.com/myuser/myrepo git:gitlab.com/mygitlabuser/mirrorrepo --origin-ssh-key /path/to/origin_key --destination-ssh-key /path/to/destination_key
```
> This will clone the origin repository into origin.git local repository, then push it to the destination repository.

- Define a directory for the local clone of the repository:
```console
gitmirror sync git:github.com/myuser/myrepo git:gitlab.com/mygitlabuser/mirrorrepo --origin-ssh-key /path/to/origin_key --destination-ssh-key /path/to/destination_key --workdir /path/to/local/clone
```

- Discover repositories for a user and sync them to a destination:
```console
gitmirror --discover-origin github --discover-origin-username myuser --discover-destination gitlab --discover-destination-username mygitlabuser --origin-ssh-key /path/to/origin_key --destination-ssh-key /path/to/destination_key
```

- Using a configuration file:
```console
gitmirror --config /path/to/config.yaml
```

- Generate a summary report of the synchronization process:
```console
gitmirror sync git:github.com/myuser/myrepo git:gitlab.com/mygitlabuser/mirrorrepo --origin-ssh-key /path/to/origin_key --destination-ssh-key /path/to/destination_key --summary gitmirrorsync.json
```

- Enable debug mode for more verbose output:
```console
gitmirror sync git:github.com/myuser/myrepo git:gitlab.com/mygitlabuser/mirrorrepo --origin-ssh-key /path/to/origin_key --destination-ssh-key /path/to/destination_key --debug
```

- Remove local clones after synchronization:
```console
gitmirror sync git:github.com/myuser/myrepo git:gitlab.com/mygitlabuser/mirrorrepo --origin-ssh-key /path/to/origin_key --destination-ssh-key /path/to/destination_key --cleanup
```

## Development

### Prerequisites
- Go 1.16 or later
- Docker and Docker Compose (for containerized development)

### Installation

1. Clone the repository:
   ```
   git clone https://github.com/Desvelao/gitmirror.git
   cd gitmirror
   ```

2. Run Docker Compose to set up the development environment:
   ```
   docker compose up -d
   ```

3. Access the development container:
   ```
   docker compose exec dev bash
   ```

4. Install dependencies:
   ```
   cd src
   go mod tidy
   ```

### Running the Application

To run the application locally:
```
go run ./cli --help
```

### Formatting

Format Go source files and the README with:

```bash
./scripts/format.sh
```

- Go files are formatted with `gofmt`.
- `README.md` is formatted with `prettier` (or `npx prettier` if `prettier` is not installed globally).

## Build

To build the application:
```
go build -o gitmirror ./cli
```

or 

```bash
./scripts/build.sh
```

## Release

Releases are automated with GoReleaser and GitHub Actions via [.github/workflows/release.yml](.github/workflows/release.yml).

1. Ensure your changes are merged to the default branch.
2. Create and push a semantic version tag:
   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```
3. The `Release` workflow will:
   - build binaries for Linux, macOS, and Windows
   - create archives and `checksums.txt`
   - create a GitHub Release and attach all artifacts

You can also run the workflow manually from the Actions tab using `workflow_dispatch`.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
