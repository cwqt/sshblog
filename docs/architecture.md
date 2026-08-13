# sshblog — architecture

A terminal blog served over SSH: it reads a directory of markdown posts and
renders them as a Bubble Tea TUI, styled with lipgloss + glamour and served via
Charm's Wish. All user-facing text and the posts directory come from
`sshblog.yaml`.

```
  ┌──────────────┐        just dev
  │   justfile   │  ─ go build sshblog ─►  ./sshblog >/tmp/sshblog.log 2>&1 &
  │              │        then: ssh -tt localhost -p 2222 ───────────────┐
  └──────────────┘                                                       │
                                                                         │ SSH (:2222)
                                                                         ▼
  ┌─────────────────────────────────────────────────────────────────────────────────┐
  │  main.go                                                                        │
  │  ───────                                                                        │
  │   • cfg   := config.Load("sshblog.yaml")   (defaults if file/field missing)      │
  │   • posts := post.LoadAll(cfg.Posts)    (once, at startup)                      │
  │   • wish.NewServer(WithHostKeyPath(.ssh/id_ed25519),                            │
  │                    WithMiddleware(bubbletea.Middleware(teaHandler(posts, cfg))))│
  │                                                                                 │
  │   teaHandler(posts, cfg)  ─ per SSH session ─►  renderer := MakeRenderer(s)     │
  │        └──► ui.New(posts, renderer, cfg)        (reads CLIENT pty TERM + bg)    │
  └───────────────────────────────────────┬─────────────────────────────────────────┘
                                          │
        ┌───────────────────┬─────────────┴─────────────────┐
        ▼                   ▼                               ▼
  ┌──────────────┐   ┌──────────────┐             ┌───────────────────────────────┐
  │ internal/    │   │ internal/    │             │  internal/ui (Bubble Tea MVU) │
  │   config     │   │   post       │             │  ───────────────────────────  │
  │  ─────────   │   │  ─────────   │             │  Model{Posts, cfg, renderer,  │
  │  Load()→cfg  │   │  LoadAll(dir)│  []post     │        styles, Page, Cursor…} │
  │  Title,      │   │  parses YAML │  ─────────► │                               │
  │  Description │   │  front-matter│             │  Update: WindowSize / KeyMsg  │
  │  Help, Posts │   │  → []Post    │             │  View:   index ◂▸ reader      │
  └──────────────┘   └──────────────┘             └───────────────┬───────────────┘
        │                                                         │
        │  cfg.Title / cfg.Description / cfg.Help                 ▼
        └───────────────────────────────────┐    ┌──────────────────────────────────┐
                                            ▼    │  index.go   reader.go            │
                                    ┌────────────┴──┐  · index: logo+desc+list      │
                                    │  styles.go    │  · reader: renderMarkdown()   │
                                    │  ───────────  │      collapse soft breaks →   │
                                    │  newStyles(r) │      glamour (no wrap/margin) │
                                    │  Oxocarbon    │      ansi.Wrap+Hardwrap ≤80   │
                                    └───────┬───────┘                               │
                                            └───────────────────────────────────────┘
                                            │
                                            ▼
                         ┌────────────────────────────────┐
                         │ lipgloss.Renderer (per session) │  ← bound to the CLIENT
                         │ + glamour                       │    terminal, not the
                         │ emit ANSI/SGR color codes ──────┼──► server's stdout
                         └────────────────────────────────┘
                                            │
                                            ▼
                              back over SSH to the client's terminal
```

## The flows

1. **Startup (once):** `main.go` loads `sshblog.yaml` via `config.Load` (falling
   back to built-in defaults for a missing file or empty field), then loads all
   posts from `cfg.Posts` and starts the Wish server.

2. **Per connection:** each SSH session gets its **own** `lipgloss.Renderer`
   from `bubbletea.MakeRenderer(s)`, threaded — together with `cfg` — into
   `Model` → `Styles` → glamour. This is the seam the color handling lives on:
   every color decision reads the _client's_ terminal, never the server
   process's stdout. (Using the global renderer instead strips all color under
   `just dev`, where the server's stdout is a log file.)

3. **Runtime (MVU loop):** Bubble Tea drives `Update`/`View`. State is `Page`
   (Index ↔ Reader) plus `ScrollOffset`; `index.go` and `reader.go` are the two
   pages, rendering through `m.styles.*`, `m.cfg.*`, and `m.renderMarkdown`.
   Markdown is collapsed (soft breaks joined), styled by glamour without its own
   wrap/margin, then wrapped with `ansi.Wrap` + `ansi.Hardwrap` to a hard
   80-column cap.

4. **The index list (`list.go`):** the post index is a bubbles/list
   (`list.Model`), giving filtering (`/`), pagination, and the help footer for
   free. `updateIndex` forwards keys to the list except `enter` (open the
   selected post) and `r` (open the feed), and hands everything to the list
   while its filter input is focused. Crucially, `newPostList` rebuilds **every**
   style (list, delegate, help) from the session renderer — bubbles' package
   defaults use the global renderer, which would render the list colorless under
   `just dev`, where the server's stdout is a log file. The logo + description
   are still drawn by `index.go` above the list (the list's own title bar is
   hidden).

5. **RSS feed (build-time generation):** the feed is *not* generated by the
   running server. `sshblog -gen-feed` (see `main.go`) loads the config + posts,
   calls `feed.Generate(posts, cfg)`, writes the RSS 2.0 document to
   `cfg.FeedPath`, and exits. This is meant to run at build/deploy time with
   `feed_path` pointing inside a co-located web server's document root, so the
   feed is served over HTTP at a stable public URL (e.g.
   `https://cass.si/rss.xml`) rather than by sshblog itself. When `cfg.URL` is
   set, the `<channel>` link and item `<link>`/`<guid>` permalinks are absolute
   `<url>/<slug>`; otherwise items fall back to slug-only, non-permalink guids.

6. **The `r` key:** `ui.New` resolves a `feedTarget` from config —
   `<url>/<basename(feed_path)>` when a URL is set, else the local `feed_path`
   file. `r` (shown as `r rss` in the list's help footer) fires `open.go`, which
   hands that target to the OS opener (`open`/`xdg-open`/`start`). The opener
   runs on the **server** host, so it only reaches a browser when the server
   shares a desktop with the client (local `just dev`); a launch failure is
   surfaced on a footer line under the list (`index.go`'s `openErr`, which
   `resizeList` accounts for so the frame never overflows the alt screen).
   Remote readers instead point their feed reader at the public URL directly.
