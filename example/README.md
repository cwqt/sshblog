# Example: deploying sshblog on Fly.io

A self-contained reference deployment. One Fly.io machine runs two processes
side by side under `supervisord`:

- **sshblog** — serves the blog over SSH (internal `:2222`, exposed as `:22`).
- **nginx** — serves a static landing page and the generated `rss.xml` over
  HTTP (internal `:8080`, fronted by Fly's TLS).

```
                       ┌───────────────────────── Fly machine ─────────────────────────┐
  ssh example.com ──►  :22  ──────────────────────────────────►  sshblog  (:2222)      │
                       │                                            │ reads posts/,    │
                       │                                            │ sshblog.yaml,    │
  https://example.com ─┼─► Fly TLS ─► :8080 ─► nginx ─┐             │ .ssh/ (host key) │
                       │                              ├─ index.html │                  │
                       │                              └─ rss.xml ◄──┘ written at build │
                       └───────────────────────────────────────────────────────────────┘
```

## Layout

```
sshblog.yaml        # title, description, posts dir, url, feed_path
posts/              # YYYY-MM-DD-slug.md, with title/date front-matter
deploy/
  index.html        # HTTP landing page (points visitors at the ssh command)
  nginx.conf        # serves index.html + rss.xml
  supervisord.conf  # runs sshblog + nginx together
Dockerfile          # go install sshblog, generate rss.xml, run both
fly.toml            # HTTP :8080 + SSH :22->2222, host-key volume
```

## How the RSS feed works

The feed is generated **at build time**. The
`Dockerfile` runs `sshblog -gen-feed`, which reads `sshblog.yaml` + `posts/`
and writes the RSS document to `feed_path` (`/var/www/html/rss.xml`, inside
nginx's document root). nginx then serves it at `https://<your-domain>/rss.xml`,
and pressing `r` in the TUI opens that URL.

## Deploy

1. Install [flyctl](https://fly.io/docs/flyctl/install/) and `fly auth login`.

2. Edit the placeholders:

   - `fly.toml` — set `app` to a unique name and `primary_region` to your
     nearest [region](https://fly.io/docs/reference/regions/).
   - `sshblog.yaml` — set `url` to your domain (drives RSS permalinks and the
     address `r` opens).
   - `deploy/index.html` — update the `ssh example.com` command.

3. Create the app and the volume that persists the SSH host key (so returning
   visitors don't get "host key changed" warnings):

   ```sh
   fly apps create <your-app-name>
   fly volumes create hostkey --size 1
   ```

4. Deploy:

   ```sh
   fly deploy
   ```

5. Point your domain at the app and connect:

   ```sh
   ssh <your-domain>
   ```

> **Note:** `Dockerfile` uses `go install github.com/cwqt/sshblog@latest`. Pin a
> released tag (`@v0.2.0`) for reproducible builds. `-gen-feed`, `url`, and
> `feed_path` require sshblog ≥ the release that introduced them.

## Local preview

You don't need Fly to see the blog. From this directory:

```sh
go install github.com/cwqt/sshblog@latest
sshblog                                     # listens on :2222
ssh -p 2222 localhost \
  -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null
```

To preview the generated feed locally, write it somewhere writable and open it:

```sh
sshblog -gen-feed   # writes to feed_path from sshblog.yaml
```
