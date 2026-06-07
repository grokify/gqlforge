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
gqlforge introspect
      ↓
schema.graphql
      +
operations/           ← You define these
  users.graphql
  projects.graphql
  search.graphql
      ↓
gqlforge generate (wraps genqlient)
      ↓
generated/client.go
      ↓
SDK facade (optional wrapper)
```

## Features

- **Introspection** - Fetch GraphQL schemas from any endpoint
- **SDL Export** - Convert introspection results to Schema Definition Language
- **Authentication** - Support for bearer tokens and goauth credentials
- **Code Generation** - Generate type-safe Go clients via genqlient
- **Scaffold** - Generate operation stubs from schema (coming soon)

## Installation

```bash
go install github.com/grokify/gqlforge/cmd/gqlforge@latest
```

## Quick Start

```bash
# 1. Initialize project structure
gqlforge init myapi

# 2. Introspect the GraphQL API
gqlforge introspect --token YOUR_TOKEN https://api.example.com/graphql

# 3. Define your operations (queries/mutations)
cat > operations/users.graphql << 'EOF'
query GetUser($id: ID!) {
  user(id: $id) {
    id
    name
    email
  }
}

query ListUsers($first: Int) {
  users(first: $first) {
    id
    name
  }
}
EOF

# 4. Generate Go client
gqlforge generate
```

## Commands

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

Global Flags:
      --account string   Account key in credentials file
      --creds string     Path to goauth credentials file
  -o, --output string    Output directory (default ".")
  -v, --verbose          Enable verbose output
```

### Examples

```bash
# Basic introspection
gqlforge introspect https://api.github.com/graphql --token ghp_xxx

# Using goauth credentials
gqlforge introspect --creds ~/.config/goauth/creds.json --account github \
    https://api.github.com/graphql

# Using environment variables
export GQLFORGE_TOKEN=your_token
export GQLFORGE_ENDPOINT=https://api.example.com/graphql
gqlforge introspect
```

## Output Formats

### SDL (Schema Definition Language)

The default output format. Human-readable and compatible with genqlient.

```graphql
type Query {
  user(id: ID!): User
  users: [User!]!
}

type User {
  id: ID!
  name: String!
  email: String
}
```

### JSON (Introspection Result)

Raw introspection result for tooling and programmatic use.

```json
{
  "schema": {
    "queryType": { "name": "Query" },
    "types": [...]
  }
}
```

## Roadmap

- [ ] `generate` command - Generate Go client using genqlient
- [ ] `init` command - Initialize a new project with config
- [ ] Full goauth integration
- [ ] Custom scalar mappings
- [ ] Operation file generation

## License

MIT
