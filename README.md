# sshblog

Host your blog over SSH.

## Quick start

```sh
just dev      # build, start the server, and connect over SSH
```

Or manually:

```sh
just build    # produces ./sshblog
./sshblog     # listens on 0.0.0.0:2222
ssh -p 2222 localhost
```

Recipes are defined in the [`justfile`](justfile); run `just` to list them.
Install with `brew install just`.

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

| Key                 | Action                         |
| ------------------- | ------------------------------ |
| `↑`/`k`, `↓`/`j`    | move / scroll                  |
| `/`                 | filter posts                   |
| `enter`             | open the selected post         |
| `r`                 | open the RSS feed in a browser |
| `esc` / `backspace` | back to the list               |
| `?`                 | toggle full help               |
| `q` / `ctrl+c`      | quit                           |

The post index is a [bubbles `list`](https://github.com/charmbracelet/bubbles#list),
so it comes with fuzzy filtering (`/`) and pagination for free.

## RSS

The RSS 2.0 feed is generated ahead of time and served over HTTP by a web
server sitting alongside sshblog, so it lives at a stable public URL. Two config
fields drive it:

```yaml
url: https://example.com          # public base URL, used for feed permalinks
feed_path: /var/www/html/rss.xml  # where `sshblog -gen-feed` writes the feed
```

Generate the feed with:

```sh
sshblog -gen-feed   # writes the RSS document to feed_path, then exits
```

Run this at build/deploy time and point `feed_path` inside your web server's
document root. With `url` set, the feed is then reachable at
`<url>/<basename of feed_path>` (e.g. `https://example.com/rss.xml`), and feed
items carry absolute `<url>/<slug>` permalinks. Without `url` the feed still
generates, with slug-only item identifiers.

Pressing `r` (shown as `r rss` in the footer help) hands that address to the
system's default handler via `open`/`xdg-open`/`start`.

> **Note:** the opener runs on the machine hosting the SSH server, not the
> client, so it only surfaces a browser during local development
> (`ssh localhost`, where server and client share a desktop). Readers on a
> remote deployment instead point their feed reader straight at the public URL.

## Deployment

See [`example/`](example/) for a complete Fly.io deployment: a single machine
running sshblog (SSH) and nginx (HTTP landing page + `rss.xml`) side by side
under supervisord.

See [`docs/architecture.md`](docs/architecture.md) for further details.
