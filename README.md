# fluxstigmergy

Stigmergic communication for Go — agents leave traces in a shared environment, enabling coordination through the environment itself rather than direct messaging.

## Core Concept

Inspired by ant colonies and stigmergy in nature: agents deposit, read, and modify traces in a shared `SharedEnvironment`. Traces decay over time (halflife-based), are boosted by reads, and are garbage-collected when weak.

## Usage

```go
env := fluxstigmergy.NewEnv()

// Agent 1 leaves a trace
env.Deposit(1, "food:nest", "at north entrance", fluxstigmergy.InfoTrace, 500)

// Agent 2 reads it
tr := env.Read("food:nest")
fmt.Println(tr.Value) // "at north entrance"

// Decay over time — traces lose strength, reads provide boost
removed := env.Decay(1*time.Hour, 5, 10) // removes traces below strength 10
```

## API

| Method | Description |
|---|---|
| `Deposit(author, key, value, type, strength)` | Leave a trace, returns index |
| `Read(key)` | Read a trace by key (increments read count) |
| `ReadAll(prefix, max)` | Read traces matching key prefix |
| `Modify(author, key, newValue, strengthAdd)` | Update a trace (author only) |
| `Erase(author, key)` | Remove a trace (author only) |
| `Decay(halfLife, readBoost, minStrength)` | Apply decay, remove weak traces |
| `ByAuthor(author)` | Filter traces by author |
| `ByType(type)` | Filter traces by type |
| `Strongest(n)` | Top-n traces by strength |
| `Oldest(n)` | Top-n oldest traces |
| `Stats(minStrength)` | Aggregate statistics |

## Trace Types

- `InfoTrace` — informational markers
- `WarningTrace` — warnings/dangers
- `ClaimTrace` — resource claims
- `WaypointTrace` — navigation markers
- `BoundaryTrace` — territorial boundaries

## License

MIT
