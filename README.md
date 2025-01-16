# Git Identity Manager (gitid)

A terminal-based tool that helps developers manage multiple Git identities easily through an interactive interface.

![demo](https://github.com/user-attachments/assets/8ec86e59-2cb1-47b7-9acd-54a7d0d8f20f)

## Features

- 🔄 Switch between multiple Git identities with ease
- ➕ Add new identities interactively
- 🗑️ Delete unwanted identities
- 💻 Terminal-based UI with keyboard navigation
- 🔒 Uses Git's built-in configuration system

## Installation


### From Binary

Download the appropriate binary for your platform from the [releases page](https://github.com/nathabonfim59/gitid/releases).

### From Package (Linux)

#### Debian/Ubuntu:
```bash
sudo dpkg -i gitid_*.deb
```

#### RedHat/Fedora:
```bash
sudo rpm -i gitid_*.rpm
```

### Building from Source

#### Prerequisites

- Go 1.21 or later
- Make
- NFPM (for package generation)

#### Build Commands

```bash
# Build for your current platform
make build

# Build static binary (Linux only)
make build-static

# Create releases for all platforms and packages
make release

# Clean build artifacts
make clean
```

## Usage

Run `gitid` to start the interactive interface.

### Keyboard Controls

- `↑`/`↓` or `j`/`k` - Navigate through identities
- `Enter` - Select identity or confirm action
- `D` - Delete selected identity
- `←`/`→` - Navigate confirmation dialog
- `Esc` - Cancel current action
- `q` - Quit application

### Managing Identities

- **Switch Identity**: Select an identity from the list and press Enter
- **Add Identity**: Select "Add new identity" and follow the prompts
- **Delete Identity**: Navigate to an identity and press D, then confirm

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
