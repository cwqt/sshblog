# sshblog

Host your blog over SSH.

## Quick start

```sh
make dev      # build, start the server, and connect over SSH
```

Or manually:

```sh
make build    # produces ./sshblog
./sshblog     # listens on 0.0.0.0:2222
ssh -p 2222 localhost
```

## Posts

Drop markdown files with YAML front-matter into the posts directory (default
`posts/`):

```markdown
---
title: Hello, world
date: 2024-01-01T00:00:00Z
---

Your post body in **markdown**.
```

The filename's `YYYY-MM-DD-slug.md` prefix supplies the slug; the title and date
come from the front-matter. Posts are listed newest-first.

## Configuration

All user-facing text and the posts directory come from `sshblog.yaml`. Every field
is optional and falls back to a built-in default:

```yaml
title: sshblog
description: |-
  A little blog you read over SSH.
posts: posts
help:
  index: "↑/k up • ↓/j down • enter select • q quit"
  reader: "↑/k up • ↓/j down • esc/backspace back • q quit"
```

## Keys

| Key                 | Action           |
| ------------------- | ---------------- |
| `↑`/`k`, `↓`/`j`    | move / scroll    |
| `enter`             | open a post      |
| `esc` / `backspace` | back to the list |
| `q` / `ctrl+c`      | quit             |

See [`docs/architecture.md`](docs/architecture.md) for further details.
