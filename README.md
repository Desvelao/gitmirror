# gitmirror

## Overview

`gitmirror` is a command-line interface (CLI) application designed to synchronize git repositories. It provides functionalities for discovering repositories, mirroring them, and keeping them in sync between the two platforms.

Modes:
- `origin-destination`: Synchronize changes between origin and destination.
- `config`: Use a configuration file to specify repositories and settings.
- `discover`: Search for repositories in a provider and sync them to a destination.
> Only GitHub

## Installation

```console
wget <RELEASE_TAR_GZ_FILE_URL> \
&& tar -xzf <TAR_GZ_FILE> \
&& sudo mv gitmirror /usr/local/bin/
```

See [Releases](https://github.com/Desvelao/gitmirror/releases) for the latest version and installation instructions for other platforms.

For example:
```console
wget https://github.com/Desvelao/gitmirror/releases/download/v0.0.1-alpha3/gitmirror-0.0.1-alpha3-linux-amd64.tar.gz \
&& tar -xzf gitmirror-0.0.1-alpha3-linux-amd64.tar.gz \
&& sudo mv gitmirror /usr/local/bin/
```

Check installation:
```console
gitmirror --version
```

## Usage

### Syncing Repositories

#### Mode: origin-destination

- Specify origin and destination repositories with SSH keys:
```console
gitmirror sync git:github.com/myuser/myrepo.git git:gitlab.com/mygitlabuser/mirrorrepo.git --origin-ssh-key /path/to/origin_key --destination-ssh-key /path/to/destination_key
```
> This will clone the origin repository into origin.git local repository, then push it to the destination repository.

- Specify a local clone directory for the synchronization process:
```console
gitmirror sync git:github.com/myuser/myrepo.git git:gitlab.com/mygitlabuser/mirrorrepo.git --origin-ssh-key /path/to/origin_key --destination-ssh-key /path/to/destination_key --clone-dir /path/to/local/clone
```

The local clone directory will be used to store the temporary clones of the repositories during the synchronization process. If not specified, it defaults to the current working directory.

- Specify SSH options for the origin and destination repositories:
```console
gitmirror sync git:github.com/myuser/myrepo.git git:gitlab.com/mygitlabuser/mirrorrepo.git --origin-ssh-key /path/to/origin_key --destination-ssh-key /path/to/destination_key --origin-ssh-options "-o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no" --destination-ssh-options "-o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no"
```

### Mode: discover

- Discover repositories for a user and sync them to a destination:
```console
gitmirror sync --discover-origin github --discover-origin-username myuser --discover-destination gitlab --discover-destination-username mygitlabuser --origin-ssh-key /path/to/origin_key --destination-ssh-key /path/to/destination_key
```

### Mode: config

- Using a configuration file:
```console
gitmirror sync --config /path/to/config.yaml
```

### Additional Options

- Generate a summary report of the synchronization process:
```console
gitmirror sync git:github.com/myuser/myrepo.git git:gitlab.com/mygitlabuser/mirrorrepo.git --origin-ssh-key /path/to/origin_key --destination-ssh-key /path/to/destination_key --summary gitmirrorsync.json
```

- Enable debug mode for more verbose output:
```console
gitmirror sync git:github.com/myuser/myrepo.git git:gitlab.com/mygitlabuser/mirrorrepo.git --origin-ssh-key /path/to/origin_key --destination-ssh-key /path/to/destination_key --debug
```

- Remove local clones after synchronization:
```console
gitmirror sync git:github.com/myuser/myrepo.git git:gitlab.com/mygitlabuser/mirrorrepo.git --origin-ssh-key /path/to/origin_key --destination-ssh-key /path/to/destination_key --cleanup
gitmirror sync git:github.com/myuser/myrepo.git git:gitlab.com/mygitlabuser/mirrorrepo.git --origin-ssh-key /path/to/origin_key --destination-ssh-key /path/to/destination_key --cleanup --clone-dir /path/to/local/clone
```

## Usage as systemd service

To run `gitmirror` as a systemd service, you can create a service unit file. Below is an example of how to set this up:

1. Create a systemd service file, e.g., `/etc/systemd/system/gitmirror-sync.service` with the following content:

```
[Unit]
Description=GitMirror Sync Service
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot
User=myuser
ExecStart=/usr/local/bin/gitmirror sync --discover-origin github --discover-destination gitlab --discover-origin-username MyUsername --discover-destination-username DestinationUsername --destination-ssh-key /path/to/keys/destination --origin-ssh-key /path/to/keys/origin --destination-ssh-options '-o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no' --origin-ssh-options '-o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no' -w /path/to/clone/dir --cleanup

[Install]
WantedBy=multi-user.target
```

> Replace the user to run the service.

2. Create a timer unit file, e.g., `/etc/systemd/system/gitmirror-sync.timer` with the following content:

```
[Unit]
Description=Run GitMirror sync every Sunday at midnight

[Timer]
OnCalendar=Sun 00:00
Persistent=true

[Install]
WantedBy=timers.target
```

3. Reload systemd to recognize the new service and timer:
```bash
sudo systemctl daemon-reload
sudo systemctl enable --now gitmirror-sync.timer
```

4. Check the status of the timer and service:
```bash
sudo systemctl status gitmirror-sync.service
sudo systemctl status gitmirror-sync.timer
systemctl list-timers --all gitmirror.timer
```

Optionally, you can also run the service manually to test it:
```bash
sudo systemctl start gitmirror-sync.service
```

## Development

### Prerequisites
- Go 1.18 or later
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
2. Update installation command version in the README:
   ```bash
   ./scripts/bump-version.sh v1.2.3
   ```
3. Create and push a semantic version tag:
   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```
4. The `Release` workflow will:
   - build binaries for Linux, macOS, and Windows
   - create archives and `checksums.txt`
   - create a GitHub Release and attach all artifacts

You can also run the workflow manually from the Actions tab using `workflow_dispatch`.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any enhancements or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for details.
