# CLI Reference

## Global Flags

| Flag | Description |
|------|-------------|
| `--creds <file>` | Path to goauth credentials file |
| `--account <key>` | Account key in credentials file |
| `-o, --output <dir>` | Output directory (default: `.`) |
| `-v, --verbose` | Enable verbose output |

## Commands

### introspect

Fetch a GraphQL schema via introspection.

```bash
gqlforge introspect [endpoint] [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--endpoint <url>` | GraphQL endpoint (alternative to positional arg) |
| `--token <token>` | Bearer token for authentication |
| `--sdl` | Output schema as SDL (.graphql) (default: true) |
| `--json` | Output raw introspection result (.json) |
| `--name <name>` | Base name for output files (default: `schema`) |

**Examples:**

```bash
# Basic introspection
gqlforge introspect https://api.github.com/graphql --token ghp_xxx

# Using goauth credentials
gqlforge introspect --creds ~/.config/goauth/creds.json --account github \
    https://api.github.com/graphql

# Output both SDL and JSON
gqlforge introspect --sdl --json https://api.example.com/graphql
```

### init

Initialize a new gqlforge project.

```bash
gqlforge init <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--endpoint <url>` | GraphQL endpoint URL (saved to config) |
| `--package <name>` | Go package name for generated code (default: `graphql`) |

**Creates:**

```
<name>/
├── genqlient.yaml        # genqlient configuration
├── schema.graphql        # GraphQL schema (run introspect to populate)
├── operations/           # Your GraphQL operations
│   └── example.graphql   # Example operation file
├── generated/            # Generated Go code (after running generate)
├── gqlforge.yaml         # Project configuration
└── .gitignore
```

**Example:**

```bash
gqlforge init myapi --package myapi --endpoint https://api.example.com/graphql
```

### generate

Generate Go client code using genqlient.

```bash
gqlforge generate [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--config <file>` | Path to genqlient.yaml (default: `genqlient.yaml`) |

**Prerequisites:**

1. Install genqlient: `go install github.com/Khan/genqlient@latest`
2. Have a valid `schema.graphql`
3. Define operations in `operations/*.graphql`
4. Configure `genqlient.yaml`

### version

Print version information.

```bash
gqlforge version
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GQLFORGE_TOKEN` | Bearer token for authentication |
| `GQLFORGE_ENDPOINT` | Default GraphQL endpoint URL |
