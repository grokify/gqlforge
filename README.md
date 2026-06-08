# GQLForge

[![Go CI][go-ci-svg]][go-ci-url]
[![Go Lint][go-lint-svg]][go-lint-url]
[![Go SAST][go-sast-svg]][go-sast-url]
[![Go Report Card][goreport-svg]][goreport-url]
[![Docs][docs-godoc-svg]][docs-godoc-url]
[![Visualization][viz-svg]][viz-url]
[![License][license-svg]][license-url]

 [go-ci-svg]: https://github.com/grokify/gqlforge/actions/workflows/go-ci.yaml/badge.svg?branch=main
 [go-ci-url]: https://github.com/grokify/gqlforge/actions/workflows/go-ci.yaml
 [go-lint-svg]: https://github.com/grokify/gqlforge/actions/workflows/go-lint.yaml/badge.svg?branch=main
 [go-lint-url]: https://github.com/grokify/gqlforge/actions/workflows/go-lint.yaml
 [go-sast-svg]: https://github.com/grokify/gqlforge/actions/workflows/go-sast-codeql.yaml/badge.svg?branch=main
 [go-sast-url]: https://github.com/grokify/gqlforge/actions/workflows/go-sast-codeql.yaml
 [goreport-svg]: https://goreportcard.com/badge/github.com/grokify/gqlforge
 [goreport-url]: https://goreportcard.com/report/github.com/grokify/gqlforge
 [docs-godoc-svg]: https://pkg.go.dev/badge/github.com/grokify/gqlforge
 [docs-godoc-url]: https://pkg.go.dev/github.com/grokify/gqlforge
 [viz-svg]: https://img.shields.io/badge/visualization-Go-blue.svg
 [viz-url]: https://mango-dune-07a8b7110.1.azurestaticapps.net/?repo=grokify%2Fgqlforge
 [loc-svg]: https://tokei.rs/b1/github/grokify/gqlforge
 [repo-url]: https://github.com/grokify/gqlforge
 [license-svg]: https://img.shields.io/badge/license-MIT-blue.svg
 [license-url]: https://github.com/grokify/gqlforge/blob/main/LICENSE

GraphQL client SDK generator for Go.

`gqlforge` orchestrates [genqlient](https://github.com/Khan/genqlient) to generate type-safe Go client code from GraphQL schemas and operations.

## Key Concepts

Unlike OpenAPI (where the spec defines all operations), GraphQL requires **two artifacts**:

1. **Schema** - The GraphQL type system (fetched via introspection)
2. **Operations** - The specific queries/mutations your SDK exposes (user-defined)

```
GraphQL Endpoint
      ↓
gqlforge introspect      ← Fetch schema
      ↓
schema.graphql
      ↓
gqlforge scaffold        ← Generate stub operations
      ↓
operations/*.graphql     ← Edit generated stubs
      ↓
gqlforge validate        ← Check before generation
      ↓
gqlforge generate        ← Generate Go client (embedded genqlient)
      ↓
generated/client.go

# For ongoing development:
gqlforge watch           ← Auto-regenerate on changes
gqlforge diff            ← Detect schema drift
```

## Features

- **Introspection** - Fetch GraphQL schemas from any endpoint
- **SDL Export** - Convert introspection results to Schema Definition Language
- **Scaffold** - Generate stub operations from Query/Mutation types
- **Validate** - Check operations against schema before generation
- **Diff** - Compare local schema to remote endpoint (detect drift)
- **Watch** - Auto-regenerate on file changes
- **Code Generation** - Generate type-safe Go clients via embedded genqlient
- **Authentication** - Support for bearer tokens and goauth credentials

## Installation

```bash
go install github.com/grokify/gqlforge/cmd/gqlforge@latest
```

## Quick Start

```bash
# 1. Initialize project structure
gqlforge init myapi
cd myapi

# 2. Introspect the GraphQL API
gqlforge introspect --token YOUR_TOKEN https://api.example.com/graphql

# 3. Generate stub operations from schema
gqlforge scaffold schema.graphql

# 4. Edit the generated operations as needed
# operations/queries.graphql, operations/mutations.graphql

# 5. Validate operations against schema
gqlforge validate

# 6. Generate Go client
gqlforge generate
```

## Commands

| Command | Description |
|---------|-------------|
| `introspect` | Fetch GraphQL schema via introspection query |
| `init` | Initialize genqlient project structure |
| `scaffold` | Generate stub operations from schema types |
| `validate` | Validate operations against schema |
| `diff` | Compare local schema to remote endpoint |
| `watch` | Watch files and auto-regenerate on changes |
| `generate` | Generate Go client code using genqlient |
| `version` | Display version information |

### introspect

Fetch a GraphQL schema via introspection.

```bash
gqlforge introspect [endpoint] [flags]

Flags:
      --endpoint string   GraphQL endpoint (alternative to positional arg)
      --json              Output raw introspection result (.json)
      --name string       Base name for output files (default "schema")
      --sdl               Output schema as SDL (.graphql) (default true)
      --token string      Bearer token for authentication
```

### scaffold

Generate stub GraphQL operations from schema types.

```bash
gqlforge scaffold <schema.graphql> [flags]

Flags:
      --type string       Type to scaffold: Query, Mutation, or both (default "both")
      --depth int         Max field depth for selection sets (default 2)
      --include string    Glob pattern for fields to include
      --exclude string    Glob pattern for fields to exclude
  -o, --output string     Output directory (default "operations")
```

### validate

Validate GraphQL operations against schema.

```bash
gqlforge validate [flags]

Flags:
      --schema string       Path to schema file (default "schema.graphql")
      --operations string   Glob pattern for operations (default "operations/*.graphql")
      --strict              Treat warnings as errors
      --json                Output validation results as JSON
```

### diff

Compare local schema to remote endpoint.

```bash
gqlforge diff <local-schema> <remote-endpoint> [flags]

Flags:
      --token string    Bearer token for remote introspection
      --breaking-only   Only show breaking changes
      --json            Output diff as JSON
```

### watch

Watch files and auto-regenerate on changes.

```bash
gqlforge watch [flags]

Flags:
      --config string       Path to genqlient.yaml (default "genqlient.yaml")
      --debounce duration   Debounce duration (default 300ms)
```

## Global Flags

```
      --account string   Account key in credentials file
      --creds string     Path to goauth credentials file
  -o, --output string    Output directory (default ".")
  -v, --verbose          Enable verbose output
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `GQLFORGE_TOKEN` | Bearer token for authentication |
| `GQLFORGE_ENDPOINT` | Default GraphQL endpoint URL |

## Examples

```bash
# Basic introspection
gqlforge introspect https://api.github.com/graphql --token ghp_xxx

# Using goauth credentials
gqlforge introspect --creds ~/.config/goauth/creds.json --account github \
    https://api.github.com/graphql

# Scaffold only query operations
gqlforge scaffold schema.graphql --type Query -o operations/

# Validate before generation
gqlforge validate --strict

# Detect schema drift in CI
gqlforge diff schema.graphql https://api.example.com/graphql --breaking-only --json

# Development workflow with auto-regeneration
gqlforge watch
```

## License

MIT
