# gqlforge

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
      ↓
gqlforge generate (wraps genqlient)
      ↓
generated/client.go
```

## Installation

```bash
go install github.com/grokify/gqlforge/cmd/gqlforge@latest
```

## Quick Start

```bash
# 1. Initialize project structure
gqlforge init myapi

# 2. Introspect the GraphQL API
cd myapi
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
EOF

# 4. Generate Go client
gqlforge generate
```

## Features

- **Introspection** - Fetch GraphQL schemas from any endpoint
- **SDL Export** - Convert introspection results to Schema Definition Language
- **Authentication** - Support for bearer tokens and goauth credentials
- **Code Generation** - Generate type-safe Go clients via genqlient
