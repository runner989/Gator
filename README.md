# Gator – A Tiny Console RSS Reader

Gator is a small command‑line application (written in Go) that lets you:

- create users and switch between them
- add RSS feeds and follow / unfollow them
- automatically gather new posts in the background
- browse your personal timeline of articles

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

Gator reads a tiny JSON file called `.gatorconfig.json` from the working directory.  Example:

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

### Common commands

| Command               | Example                                                        | What it does                                         |
| --------------------- |----------------------------------------------------------------|------------------------------------------------------|
| `register`            | `gator register alice`                                         | create a new user                                    |
| `login`               | `gator login alice`                                            | switch current user                                  |
| `addfeed`             | `gator addfeed "Hacker News" https://news.ycombinator.com/rss` | insert a feed *and* auto‑follow it                   |
| `agg`                 | `gator agg 1m`                                                 | start the endless collector (press `Ctrl+C` to quit) |
| `browse`              | `gator browse 5`                                               | show the 5 newest posts for the logged‑in user       |
| `follow` / `unfollow` | `gator follow https://techcrunch.com/feed/`                    | change subscriptions                                 |
| `users`               | `gator users`                                                  | list all registered users                            |
| `reset`               | `gator reset`                                                  | **danger:** truncate users, feeds, follows & posts   |

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

