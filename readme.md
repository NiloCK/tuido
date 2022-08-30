An opinionated terminal interface for efficient browsing and management of [[x]it!](https://github.com/jotaen/xit) formatted todo items. Inspired, loosely, by _Getting Things Done_ (David Allen) and informed by various entries of the self-help skills & habits genre (Charles Duhigg, James Clear, Anders Ericsson, etc).

## Features

- [x] searches the working directory recursively for [x]it! compatible items in `.xit`, `.md`, and `.txt` files
- [x] compactly displays pending todos and offers navigation between `todo` and `done`
- [x] allows for creating new items, updating existing items, and persists updates to disk
- [x] allows for filtering via `tags`
- [x] a simple pomodoro mode for timeboxed focus on individual items
- [x] one-button (`z`) progressive snooze parks items for 1,2,3,5,8,... days
- [x] progressive deterrence for adding new items

![tuidi preview](./preview.gif)

## Usage

From some directory containing `[x]it!` files / items:

```
tuido
```

### In app controls

- **?**: help
- **n**: make a new item
- slected item controls:
  - **[space]**: set status open
  - **x**, **X**: set status checked (done)
  - **s**, **~**: set status obsolete
  - **a**, **@**: set status ongoing
  - **e**: edit item text
  - **p**: enter a pomodoro session for item
  - **z**: snooze this item (set a later active date)
  - **!**: bump the `importance` modifier on this item
- **[tab]**: switch between pending and done items
- **/**: filter list by `#tags`
- **[up]**, **[down]**: navigate items
- **q**: quit

## Configuration

Tuido writes new items by default to `$HOME/.tuido/YYYY-MM-DD.xit`. To set a different write location, create file `tuido.conf` in the user config directory (`$HOME/.config` in linux, `$HOME/AppData` in windows). The write location can be a file, which will be appended to, or a directory, which whill recieve datestamped `.xit` files as in the default setting.

```
writeto=~/mysingletodolist.txt
```

```
writeto=~/todos
```

Include a `.tuido` file in individual directories to add filetypes for parsing along that subtree.

```
extensions=go,js,cpp
```

## Development

0. install go (see https://go.dev)
1. (suggested) read the in-readme tutorial for https://github.com/charmbracelet/bubbletea
2. clone repo
3. `go run .`

tuido is dogfooding. The project's `.tuido` file:

- instructs tuido to parse items in `.go` files as well as the defaults.
- instructs tuido to write new items directly to this readme

Result being that the app, runnng in test, contains a good running list of development todos & a convenient method to append to the roadmap.

## Licence

GPL

## Roadmap

- [ ] #feat make new-items repsect the filetype being written to (leading comment slashes for code files, leading bullet for readme, etc)
- [@] process #dates
  - [x] from items themselves
    - [x] from #due tags
    - [ ] according to [x]it spec
  - [x] (for creation #date) from the names of an item's source file
- [@] #ui sort items by priority [ ], age [ ], or due #dates [x]
- [ ] #feat #ui provide details / context (preview into source file) on current selected item, or quick open of an item's source location
- [ ] #feat allow plain-text fuzzy text search/filter of item body text (only tag names currently)
- [@] #feat specify / parse a format for recurring items (call mom, eat a salad)
  - [x] #repeat=durationShorthand is active
  - [@] need to mark open + postpone + set new due date on `x` done markers
- [ ] have infrastructure for managing task-specific checklist files (beach trip) #feat #ui #maybe
- [@] #feat #maybe accept command line flags or config for other file extenstions, source directories, etc
- [ ] #feat #maybe fully respect / implement the [x]it spec
- [ ] #feat respect .gitignire configs
- [ ] tag v0.0.1, produce platform builds
