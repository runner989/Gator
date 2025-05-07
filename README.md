# Gator – A Tiny Console RSS Reader

Gator is a small command‑line application (written in Go) that lets you:

- create users and switch between them
- add RSS feeds and follow / unfollow them
- automatically gather new posts in the background
- browse, sort and paginate your timeline – right from the terminal
---

## Prerequisites

| Tool | Tested Version |
| ---- | -------------- |
|      |                |

| **Go**         | ≥ 1.22 |
| -------------- |--------|
| **PostgreSQL** | ≥ 15   |

Make sure each executable (`go`, `psql`, `postgres`) is in your `$PATH`.

---

## Installing the `gator` CLI

```bash
# Pick a folder that is already on your GOPATH / GOBIN.
go install github.com/runner989/gator@latest

# Verify it worked
$ gator -h
```

The binary is now available globally as `gator`.

---

## Database setup

1. **Create a database** (and a super‑simple super‑user):
   ```bash
   $ createdb gator
   $ psql -d gator -c "CREATE USER gator PASSWORD 'gator';"
   $ psql -d gator -c "GRANT ALL ON DATABASE gator TO gator;"
   ```
2. **Run the migrations** (requires [Goose](https://github.com/pressly/goose)):
   ```bash
   goose -dir sql/migrations postgres "postgres://gator:gator@localhost:5432/gator?sslmode=disable" up
   ```

---

## Configuration file

Gator reads a tiny JSON file called `.gatorconfig.json` from the home directory.  Example:

```json
{
  "DbUrl": "postgres://gator:gator@localhost:5432/gator?sslmode=disable",
  "CurrentUser": ""
}
```

\* `DbUrl` – Postgres connection string \* `CurrentUser` – will be populated after you `login`

---

## Running Gator

```bash
$ gator <command> [args]
```

## Command reference (built‑in help)

Run `gator help` or simply `gator` with no arguments to see this list at any time.

| Command                     | Example                                                        | What it does                                                                |
|-----------------------------|----------------------------------------------------------------|-----------------------------------------------------------------------------|
| `register <name>`           | `gator register alice`                                         | create a new user                                                           |
| `login <name>`              | `gator login alice`                                            | switch current user                                                         |
| `addfeed <title> <url>`     | `gator addfeed "Hacker News" https://news.ycombinator.com/rss` | insert a feed *and* auto‑follow it                                          |
| `agg <interval>`            | `gator agg 1m`                                                 | start the endless collector (press `Ctrl+C` to quit)                        |
| `browse [flags]`            | `gator browse --limit=5 --sort=title --page=2`                 | show the 5 newest posts, sorted by title, on page 2, for the logged‑in user |
| `follow` / `unfollow` `<feed>` | `gator follow https://techcrunch.com/feed/`                    | change subscriptions                                                        |
| `users`                     | `gator users`                                                  | list all registered users                                                   |
| `reset`                     | `gator reset`                                                  | **danger:** truncate users, feeds, follows & posts                          |

---

## Quick start

```bash
# 1  Create a user and log in
$ gator register alice
$ gator login alice

# 2  Add two feeds
$ gator addfeed "Hacker News" https://news.ycombinator.com/rss
$ gator addfeed "TechCrunch"  https://techcrunch.com/feed/

# 3  Start the aggregator in another terminal
$ gator agg 1m
# (Let it run for a minute; it prints as it pulls.)

# 4  Browse your timeline – newest five posts, by time, first page
$ gator browse --limit=5 --sort=time --page=0
```
---

## Development workflow

```bash
# format
make fmt

# run linters
make lint

# run tests
make test
```

---

## License

MIT © runner989 2025
