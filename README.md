# Git Identity Manager

A terminal-based tool to manage multiple Git identities.

## Building

### Prerequisites

- Go 1.21 or later
- Make
- NFPM (for package generation)

### Build Commands

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

## Installation

### From Binary

Download the appropriate binary for your platform from the releases page.

### From Package (Linux)

#### Debian/Ubuntu:
```bash
sudo dpkg -i release/gitid_*.deb
```

#### RedHat/Fedora:
```bash
sudo rpm -i release/gitid_*.rpm
```

## Usage

Run `gitid` to start the interactive prompt.

- Use arrow keys or j/k to navigate
- Press Enter to select an identity
- Press D to delete an identity
- Press q to quit
