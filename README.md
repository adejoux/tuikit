# tuikit

Shared Bubble Tea v2 building blocks for terminal UIs, used by
[portfolio-cli](https://github.com/adejoux/portfolio-cli) and cable.

tuikit keeps the look and the shortcut grammar of these tools in one place,
so they stay consistent without copying code between repositories.

## Packages

| Package | Purpose |
| --- | --- |
| `theme` | Dark and light themes (colour roles) and the shared lipgloss styles |
| `keyhints` | One-line key hints, Crush style: `key desc • key desc • ? more` |
| `dialog` | Dialog frame, title with a ╱╱╱ gradient rule, overlay on the screen |
| `help` | Searchable shortcut dialog (`f1` / `?`), filtered by context and scope |
| `fuzzy` | Subsequence fuzzy matching and ranking for filters and pickers |
| `diff` | Line diff with context, rendered for confirmation views |
| `editor` | Open a file in `$VISUAL` / `$EDITOR`, positioned on a line when supported |
| `fields` | Text field keymaps that follow the focus rules below |
| `textutil` | Width-aware wrap, crop, truncate and padding helpers |

Applications extend `theme.Theme` with their own domain colours; the roles in
tuikit are the ones every screen needs.

## Focus rules

- While a text field, filter or picker has focus, letters, digits and `?` are
  typed, not interpreted as shortcuts.
- Readline editing keys belong to text fields and are never bound to
  application actions: `ctrl+a`, `ctrl+b`, `ctrl+d`, `ctrl+e`, `ctrl+f`,
  `ctrl+h`, `ctrl+k`, `ctrl+t`, `ctrl+u`, `ctrl+v`, `ctrl+w` and `alt+`
  combinations (`fields.IsReserved`).
- `fields.TextInputKeyMap` and `fields.TextAreaKeyMap` drop the `ctrl+n` /
  `ctrl+p` line and suggestion navigation of bubbles, so `ctrl+p` can open a
  command palette everywhere. Pickers keep `ctrl+n` / `ctrl+p` to move their
  selection.

## Development

```sh
go vet ./...
go test ./...
```

## Status

Early development. The API is not stable yet.
