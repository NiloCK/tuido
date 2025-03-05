# tuido

An opinionated terminal interface for efficient browsing and management of [x]it! formatted todo items.

## Installation

```bash
npm install -g tuido
# or as a dev dependency
npm install -D tuido
```

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

See the [full documentation](https://github.com/NiloCK/tuido) for details.

## Configuration

To set a custom write location or to parse additional file types, create a `.tuido` file in your project:

```
extensions=go,js,cpp
writeto=~/mysingletodolist.txt
```

## License

GPL
