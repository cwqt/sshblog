# sshblog — architecture

A terminal blog served over SSH: it reads a directory of markdown posts and
renders them as a Bubble Tea TUI, styled with lipgloss + glamour and served via
Charm's Wish. All user-facing text and the posts directory come from
`sshblog.yaml`.

```
  ┌──────────────┐        make dev
  │   Makefile   │  ─ go build sshblog ─►  ./sshblog >/tmp/sshblog.log 2>&1 &
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
   `make dev`, where the server's stdout is a log file.)

3. **Runtime (MVU loop):** Bubble Tea drives `Update`/`View`. State is `Page`
   (Index ↔ Reader) plus `Cursor`/`ScrollOffset`; `index.go` and `reader.go` are
   the two pages, rendering through `m.styles.*`, `m.cfg.*`, and
   `m.renderMarkdown`. Markdown is collapsed (soft breaks joined), styled by
   glamour without its own wrap/margin, then wrapped with `ansi.Wrap` +
   `ansi.Hardwrap` to a hard 80-column cap.
