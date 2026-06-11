# tuido

An opinionated terminal interface for efficient browsing and management of [x]it! formatted todo items.

This npm package downloads the appropriate platform-specific binary from GitHub releases during installation.

## Installation

```bash
npm install -g tuido
# or as a dev dependency
npm install -D tuido
```

**Note**: The package will automatically download the correct binary for your platform (Linux, macOS, Windows) during installation.

## Usage

From any directory containing [x]it! files:

```bash
tuido
```

## Features

- Searches for [x]it! compatible items in `.xit`, `.md`, and `.txt` files
- Compactly displays pending todos with navigation between `todo` and `done`
- Create new items, update existing items, and persist updates to disk
- Search/filter todos by keywords
- One-button pomodoro mode for focused work
- Progressive snooze function
- Copy items to clipboard (`c` for text, `C` for text + status + source location)

See the [full documentation](https://github.com/NiloCK/tuido) for details.

## Troubleshooting

If you encounter issues with the binary installation:

1. **Binary not found**: Try reinstalling the package with `npm uninstall tuido && npm install tuido`
2. **Permission errors**: On Unix systems, ensure the binary has execute permissions
3. **Platform not supported**: Check the [releases page](https://github.com/NiloCK/tuido/releases) for available binaries
4. **Network issues**: The installation requires internet access to download the binary from GitHub

For persistent issues, you can download the binary directly from the [GitHub releases](https://github.com/NiloCK/tuido/releases).

## Configuration

To set a custom write location or to parse additional file types, create a `.tuido` file in your project:

```
extensions=go,js,cpp
writeto=~/mysingletodolist.txt
```

## License

GPL
