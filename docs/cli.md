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

### scaffold

Generate stub GraphQL operations from schema types.

```bash
gqlforge scaffold <schema.graphql> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--type <type>` | Type to scaffold: Query, Mutation, or both (default: `both`) |
| `--depth <n>` | Max field depth for selection sets (default: `2`) |
| `--include <pattern>` | Glob pattern for fields to include (comma-separated) |
| `--exclude <pattern>` | Glob pattern for fields to exclude (comma-separated) |
| `-o, --output <dir>` | Output directory (default: `operations`) |

**Examples:**

```bash
# Generate all operations
gqlforge scaffold schema.graphql -o operations/

# Only queries with depth 3
gqlforge scaffold schema.graphql --type Query --depth 3

# Filter by field name
gqlforge scaffold schema.graphql --include "user*,feature*"
gqlforge scaffold schema.graphql --exclude "internal*"
```

### validate

Validate GraphQL operations against a schema.

```bash
gqlforge validate [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--schema <file>` | Path to schema file (default: `schema.graphql`) |
| `--operations <pattern>` | Glob pattern for operations (default: `operations/*.graphql`) |
| `--strict` | Treat warnings as errors |
| `--json` | Output validation results as JSON |

**Examples:**

```bash
# Validate with defaults
gqlforge validate

# Custom paths
gqlforge validate --schema schema.graphql --operations "queries/*.graphql"

# Strict mode for CI/CD
gqlforge validate --strict

# JSON output
gqlforge validate --json
```

### diff

Compare local schema to remote endpoint (detect drift).

```bash
gqlforge diff <local-schema> <remote-endpoint> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--token <token>` | Bearer token for remote introspection |
| `--breaking-only` | Only show breaking changes |
| `--json` | Output diff as JSON |

**Examples:**

```bash
# Basic diff
gqlforge diff schema.graphql https://api.example.com/graphql --token xxx

# Using goauth credentials
gqlforge diff schema.graphql --creds creds.json --account myapi \
    https://api.example.com/graphql

# Only breaking changes
gqlforge diff schema.graphql https://api.example.com/graphql --breaking-only

# JSON output for CI/CD
gqlforge diff schema.graphql https://api.example.com/graphql --json
```

**Output:**

```
Schema Diff: schema.graphql vs https://api.example.com/graphql

+ Added Types:
  + NewType

- Removed Types (BREAKING):
  - OldType

~ Changed Types:
  ~ User:
    + newField
    - removedField (BREAKING)

Summary: 1 added, 1 removed, 2 breaking changes
```

### watch

Watch files and auto-regenerate on changes.

```bash
gqlforge watch [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--config <file>` | Path to genqlient.yaml (default: `genqlient.yaml`) |
| `--debounce <duration>` | Debounce duration for rapid changes (default: `300ms`) |

**Examples:**

```bash
# Watch with defaults
gqlforge watch

# Custom config
gqlforge watch --config custom-genqlient.yaml

# Faster response
gqlforge watch --debounce 100ms
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

1. Have a valid `schema.graphql`
2. Define operations in `operations/*.graphql`
3. Configure `genqlient.yaml`

**Note:** As of v0.2.0, genqlient is embedded and does not need to be installed separately.

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
