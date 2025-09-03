# Git Identity Manager (gitid)

A terminal-based tool that helps developers manage multiple Git identities easily through an interactive interface.

![demo](https://github.com/user-attachments/assets/8ec86e59-2cb1-47b7-9acd-54a7d0d8f20f)

## Features

- 🔄 Switch between multiple Git identities globally
- 📁 Set repository-specific local identities
- 🏷️ Optional nicknames for quick identity identification
- ➕ Add new identities interactively
- 🗑️ Delete unwanted identities
- 💻 Terminal-based UI with keyboard navigation
- 🔒 Uses Git's built-in configuration system
- 🔍 Smart identity matching by nickname, name, or email
- 📋 Command-line interface for automation and scripting
- ✨ Visual indicators for current global and local identities

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

#### Arch Linux (AUR):
```bash
# Pre-compiled binary
yay -S gitid-bin

# Build from source
yay -S gitid-git
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

GitID provides both an interactive TUI and command-line interface for managing Git identities.

### Interactive Mode

Run `gitid` (without arguments) to start the interactive interface.

#### Keyboard Controls

- `↑`/`↓` or `j`/`k` - Navigate through identities
- `Enter` - Select identity to set as global
- `r` - Set selected identity as local for current repository
- `R` - Remove local identity (use global)
- `D` - Delete selected identity
- `e` - Edit nickname for selected identity
- `E` - Edit full identity (name, email, nickname)
- `←`/`→` - Navigate confirmation dialog
- `Esc` - Cancel current action
- `q` - Quit application

### Command Line Interface

GitID also provides a full CLI for automation and scripting:

#### Global Identity Management
```bash
# List all identities
gitid list

# Show current global (and local if in repo) identity
gitid current

# Switch global identity
gitid switch <nickname|name|email>
gitid use <nickname|name|email>        # Alias for switch

# Add new identity
gitid add "Full Name" "email@domain.com" [nickname]

# Delete identity
gitid delete <nickname|name|email>

# Set/update nickname
gitid nickname <identifier> <new-nickname>
```

#### Repository-Specific (Local) Identity Management
```bash
# Show current local identity for repository
gitid repo current

# Set existing identity as local for current repository
gitid repo use <nickname|name|email>

# Add new identity and set as local for current repository
gitid repo add "Full Name" "email@domain.com" [nickname]
```

### Global vs Local Identities

GitID supports both **global** and **local** (repository-specific) Git identities:

#### Global Identity
- Applied to all Git repositories by default
- Stored in Git's global configuration (`~/.gitconfig`)
- Set with `gitid switch <identifier>` or by pressing `Enter` in TUI

#### Local Identity
- Applied only to the current repository
- Stored in the repository's local `.git/config`
- Overrides global identity for that specific repository
- Set with `gitid repo use <identifier>` or by pressing `r` in TUI

#### Visual Indicators
- **TUI**: Shows repository status and `[local]` indicators
- **CLI**: `gitid current` shows both global and local when in a repository
- **CLI**: `gitid list` shows `[current local]` indicator

### Managing Identities

#### In TUI (Interactive Mode)
- **Switch Global Identity**: Select an identity and press `Enter`
- **Set Local Identity**: Select an identity and press `r` (only in git repositories)
- **Remove Local Identity**: Press `R` to fall back to global identity
- **Add Identity**: Select "Add new identity" and follow the prompts
- **Edit Identity**: Press `e` for nickname or `E` for full identity
- **Delete Identity**: Navigate to an identity and press `D`, then confirm

#### Via Command Line
- **Global**: Use `gitid switch/use/add/delete` commands
- **Local**: Use `gitid repo use/add/current` commands

### Example Workflow

```bash
# Add work identity
gitid add "John Doe" "john.doe@company.com" work

# Add personal identity
gitid add "John Doe" "john@personal.com" personal

# Set global identity to personal
gitid switch personal

# In work repository, set local identity to work
cd ~/work-project
gitid repo use work

# Check current status
gitid current
# Output:
# Global: personal (John Doe <john@personal.com>)
# Local:  work (John Doe <john.doe@company.com>)
```

### Shell Completions

GitID supports shell completions for Bash, Zsh, and Fish to provide tab-completion for commands and arguments.

#### Installation

```bash
# Install for your current shell (auto-detected)
gitid completion bash    # For Bash
gitid completion zsh     # For Zsh
gitid completion fish    # For Fish
```

#### Removal

```bash
# Remove completions
gitid completion bash -r    # Remove Bash completions
gitid completion zsh -r     # Remove Zsh completions
gitid completion fish -r    # Remove Fish completions
```

After installation, restart your shell or source your configuration file (e.g., `source ~/.bashrc` or `source ~/.zshrc`).

### Nicknames

Nicknames are optional short identifiers that make it easier to distinguish between identities:

- **Display**: Identities with nicknames show as `nickname (Name <email>)`
- **Without nicknames**: Shows as `Name <email>` (backwards compatible)
- **Adding nicknames**: Available when creating new identities or editing existing ones
- **Smart matching**: CLI supports switching by nickname, name, or email
- **Quick identification**: Especially useful when you have multiple identities with similar names

### Help and Documentation

```bash
# Show help and all available commands
gitid help

# Get command-specific usage
gitid repo          # Shows repo subcommand usage
gitid completion    # Shows completion installation usage
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
