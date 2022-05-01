A terminal interface for [x]it! formatted todo items. Inspired, loosely, by _Getting Things Done_ (David Allen).

## usage:

```
tuido
```

It:

- [ ] searches the working directory recursively for [x]it! compatible items in `.xit`, `.md`, and `.txt` files
- [ ] compactly displays pending todos and offers navigation between `todo` and `done`
- [ ] allows for updating todo status, and persists the updates to the original files
- [ ] allows for filtering via `#tags`

It doesn't yet: (roadmap)

[ ] ] process dates
[ ] [ ] from items themselves, according to [x]it spec
[ ] [ ] (for creation date) from the names of an item's source file
- [ ] sort items by priority, age, or due dates
- [ ] allow plain-text fuzzy text search/filter of item body text (only tag names)
- [ ] have a pomodoro mode for focused work on a specific item
- [ ] specify / parse a format for recurring items (call mom, eat a salad)
- [ ] parse valued tags. EG, `#age=37` is parsed as a tag titled `age=37` rather than tag `age` with value `37`.
- [ ] contain the all-important [__z__: snooze] operation to bump an item's due date
- [ ] read a config file from `~/.tuido/config`, write by default to `~/.tuido/yyyy-mm-dd.xit`
- [ ] have infrastructure for managing task-specific checklist files (beach trip)
- [ ] accept command line flags or config for other file extenstions, source directories, etc
- [ ] allow for creating new todos or editing the body text of existing ones
- [ ] fully respect / implement the [x]it spec

In app controls:

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
