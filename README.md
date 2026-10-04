# twig

Go reimplementation of the original Rust `gl` crate. Decentralized git CLI, identity management, and MCP server for Twigpine.

## Installation

```bash
go build -o twig main.go
```

Zero external dependencies. Builds with Go 1.26 or newer.

## Usage

```bash
twig <command> [arguments...]
```

### Repository list output

```bash
twig repo list                       # Human-readable table (default)
twig repo list --format table        # Explicit table output
twig repo list --json                # Complete JSON array for scripts
twig repo list --format json         # Equivalent to --json
```

The table shows name, owner, visibility, updated date, and description. It lists
all repositories visible to the caller, without filtering by owner. Empty results
print `No repositories found.` in table mode and `[]` in JSON mode.

Scripts that parsed the previous default JSON output must add `--json` or
`--format json`. Output does not change automatically when piped or redirected.
Only `table` and `json` are supported. Combining `--json` with `--format table`
is an error; combining it with `--format json` is allowed.

JSON retains the complete response, including additional metadata and exact
numeric values. Human-readable cells strip terminal control characters and are
limited to 200 characters. JSON values are not truncated. Errors go to stderr
and return a nonzero exit status.

### Environment variables

- `TWIGPINE_NODE`: Node URL (default: `https://node.gitlawb.com`). Falls back to legacy `GITLAWB_NODE` if unset.
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

CLI integration tests use a local HTTP fixture, not the public node. To exercise
a built executable, set `TWIG_TEST_BINARY` to its absolute path and run
`go test -count=1 -run TestRepoListCLI .`. Otherwise, these tests run the CLI
entrypoint in a test subprocess.

## License

GNU Affero General Public License v3.0 only (AGPL-3.0-only). See [LICENSE](LICENSE).
