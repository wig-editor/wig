# McWig (temporary name)

Hi! Welcome to my text editor page.
McWig is a modal, Vim-like text editor written in Go. I use it as my daily driver, btw.

**Note:** McWig currently supports only `.go` files. Do not edit files that are not backed up. McWig is still in an early stage of development and may damage your files.

### Features
- LSP autocomplete, goto definition, hover info
- Tree-sitter support
- Color themes (borrowed from the Helix text editor)
- Lots of bugs
- Macro support
- Something like Emacs org-mode: Open `test.txt`, place the cursor at line 15, and press `"Ctrl-C Ctrl-C"`.

This project was written as a "speed run" — not for speed in terms of time, but rather as an exercise to explore the text editor problem space without overthinking or planning ahead. It’s a quick and "dirty" implementation, so to speak.

---

# Running

```bash
make setup-runtime
make build-run
```

---

# Keybindings

Most common Vim keybindings are implemented, providing minimal but enough (for me) support for daily source code editing. See `config/config.go` for all implemented keybindings.

To get started:

| **Key**     | **Description**           |
|-------------|---------------------------|
| Space + F   | Find files in Git project  |
| Space + B   | Buffers                   |
| Ctrl-W + V  | Split window              |
| Space + `   | Toggle file               |
| Space + /   | Search text in project    |

---

# Plans

I plan to turn this "toy project" into a stable, fully-featured Vim-like text editor.

---





