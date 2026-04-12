# MAINTENANCE.md

## Testing

```sh
go test -v ./...
```

## Design Notes

- **No concurrency safety**: `SharedEnvironment` is not goroutine-safe. Wrap with a mutex if used concurrently.
- **Decay formula**: `strength = strength / 2^(elapsed/halfLife) + reads * readBoost`, capped at 1000.
- **Strength range**: 0–1000. Modify caps at 1000; decay can reduce below minStrength triggering removal.
- **O(n) operations**: All queries are linear scans. Suitable for small-medium environments (<10k traces). For larger scales, consider index maps.

## Potential Improvements

- Add sync.RWMutex for concurrent access
- Add key-based index map for O(1) lookups
- Add persistence (JSON, BoltDB)
- Add merge/sync between environments (distributed stigmergy)
