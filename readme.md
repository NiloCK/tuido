A terminal interface for efficient browsing and management of [[x]it!](https://github.com/jotaen/xit) formatted todo items. Inspired, loosely, by _Getting Things Done_ (David Allen) and various entries of the self-help skills via habits genre (Charles Duhigg, James Clear, Anders Ericsson, etc).

## Features

- [x] searches the working directory recursively for [x]it! compatible items in `.xit`, `.md`, and `.txt` files
- [x] compactly displays pending todos and offers navigation between `todo` and `done`
- [x] allows for updating todo status, and persists the updates to the original files
- [x] allows for filtering via `tags`

![tuidi preview](./preview.gif)

## Usage

From some directory containing `[x]it!` files / items:

```
tuido
```

### In app controls

- **?**: help
- item status updates:
  - **[space]**: open
  - **x**, **X**: checked (done)
  - **s**, **~**: obsolete
  - **a**, **@**: ongoing
- **[tab]**: switch between pending and done items
- **/**: filter list by `#tags`
- **[up]**, **[down]**: navigate items
- **q**: quit

## Roadmap

- [ ] process due #dates
  - [ ] from items themselves, according to [x]it spec
  - [ ] (for creation #date) from the names of an item's source file
- [ ] #ui sort items by priority, age, or due #dates
- [ ] #feat #ui provide details / context (preview into source file) on current selected item, or quick open of an item's source location
- [ ] #feat allow plain-text fuzzy text search/filter of item body text (only tag names)
- [ ] #feat have a pomodoro mode for focused work on a specific item
- [ ] #feat specify / parse a format for recurring items (call mom, eat a salad)
- [ ] #feat parse valued tags. EG, `#age=37` is parsed as a tag titled `age=37` rather than tag `age` with value `37`.
- [ ] #feat contain the all-important [__z__: snooze] operation to bump an item's due date
- [ ] #feat read a config file from `~/.tuido/config`, write by default to `~/.tuido/yyyy-mm-dd.xit`
- [ ] have infrastructure for managing task-specific checklist files (beach trip) #feat #ui #maybe
- [ ] #feat #maybe accept command line flags or config for other file extenstions, source directories, etc
- [ ] #feat #maybe allow for creating new todos or editing the body text of existing ones
- [ ] #feat #maybe fully respect / implement the [x]it spec
- [ ] #feat read #config from any encountered `.tuido` file, and apply to child directories

## Development

0. install go (see https://go.dev)
1. (suggested) read the in-readme tutorial for https://github.com/charmbracelet/bubbletea
2. clone repo
3. `go run .`

tuido is dogfooding. As a special case, when the working directory is equal to `tuido` (as it is during `go run .` from the project root), tuido parses items in `.go` files as well as the defaults. Result being that the app, runnng in test, contains a good running list of development todos.

## Licence

GPL
