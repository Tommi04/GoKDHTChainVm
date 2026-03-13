# GoKDHTChainVm

A lightweight Go service that watches a blockchain transaction directory and updates transaction file names to maintain chain links.

## What this project does

This VM process monitors a `Blockchain` folder for new or updated `.bin` files. When a new block file appears, it:

1. Reads the incoming file name.
2. Extracts the previous block hash and current block hash from the file name.
3. Finds the prior block file that still has no "next" hash in its name.
4. Renames that prior file by appending the new block hash.

This keeps the on-disk file naming convention aligned with a linked-chain model.

## File naming model

The code assumes hashes are 64 hex chars long and uses these conventions:

- `O-<hash>.bin` → genesis block without a next hash.
- `<prevHash><hash>.bin` (128 chars total before extension) → non-genesis block without a next hash.
- `<prevHash><hash><nextHash>.bin` (192 chars total before extension) → block already linked to a following block.

When a new block `<prevHash><hash>.bin` is written, the process renames the previous block by appending `<hash>` to that previous block's filename.

## Directory layout expected at runtime

The service uses OS root as base (`/` on Linux/macOS, system drive on Windows), then expects:

- `GoChain/Blockchain` (must already exist) — watched directory containing `.bin` transaction files.
- `GoChain/BlockchainLog/chainLog.log` — created/used for runtime logs.

## Requirements

- Go 1.23+
- Access permissions to create/read/write under the `GoChain` root structure.

## Build and run

```bash
go mod tidy
go build -o gokdhtchainvm .
./gokdhtchainvm
```

On start, the service:

- Initializes logging.
- Creates an `fsnotify` watcher.
- Watches `GoChain/Blockchain` for write events.
- Processes `.bin` files to perform chain-link renaming.

## Notes and limitations

- The watched `GoChain/Blockchain` directory must exist before startup.
- The process is designed as a long-running service (it blocks forever by design).
- Only `.bin` files are processed.
- Polling logic exists as commented code; active mode uses `fsnotify`.

## Development checks

```bash
go test ./...
go build ./...
```

## License

No license file is currently included in this repository.
