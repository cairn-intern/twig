# twig

Decentralized git CLI, identity management, and MCP server for Twigpine.

## Installation

```bash
go build -o twig main.go
```

Zero external dependencies. Builds with Go 1.26 or newer.

## Usage

```bash
twig <command> [arguments...]
```

### Environment variables

- `TWIGPINE_NODE`: Node URL (default: `https://node.twigpine.com`). Falls back to legacy `GITLAWB_NODE` if unset.
- `TWIGPINE_DIR`: Configuration directory (default: `~/.twigpine`). Falls back to legacy `~/.gitlawb` if present.

### Commands

| Command | Description |
| --- | --- |
| `identity` | Manage DID keypair (generate, export, show, sign) |
| `register` | Register identity with a Twigpine node |
| `whoami` | Print identity and node status |
| `repo` | Manage repositories (create, list, info, delete) |
| `clone` | Clone a Twigpine repository |
| `issue` | Manage issues (stored as git refs) |
| `pr` | Manage pull requests |
| `mcp` | Start Model Context Protocol server over stdio |
| `node` | Node status dashboard, network info, and on-chain ops |
| `name` | Register and resolve names on Base L2 |
| `doctor` | Check installation and connectivity |
| `quickstart` | Guided interactive setup |
| `init` | Zero-to-push repository setup |
| `peer` | Peer discovery and node inspection |
| `cert` | Inspect signed ref-update certificates |
| `ipfs` | IPFS pin management and object retrieval |
| `webhook` | Manage repository webhooks |
| `mirror` | Mirror external repositories |
| `sync` | Sync repositories from peers |
| `task` | Manage agent task delegation |
| `star` | Star and unstar repositories |
| `status` | Current context snapshot |
| `agent` | List and inspect registered agents |
| `profile` | Manage agent profile |
| `protect` | Branch protection rules |
| `visibility` | Path-scoped read visibility rules |
| `changelog` | Activity changelog for a repository |
| `bounty` | Manage token bounties on repositories |
| `ucan` | Delegate and verify capability tokens |
| `version` | Print version information |

## Testing

```bash
go test -v ./...
```
