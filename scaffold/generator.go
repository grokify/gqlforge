// Package scaffold generates stub GraphQL operations from schema types.
package scaffold

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/vektah/gqlparser/v2/ast"
)

// Generator generates stub GraphQL operations from a schema.
type Generator struct {
	Schema    *ast.Schema
	MaxDepth  int
	TypeNames []string // Query, Mutation
	Include   []string // Glob patterns to include
	Exclude   []string // Glob patterns to exclude
}

// NewGenerator creates a new Generator with default settings.
func NewGenerator(schema *ast.Schema) *Generator {
	return &Generator{
		Schema:    schema,
		MaxDepth:  2,
		TypeNames: []string{"Query", "Mutation"},
	}
}

// Generate generates stub operations and returns a map of filename to content.
func (g *Generator) Generate() (map[string]string, error) {
	result := make(map[string]string)

	for _, typeName := range g.TypeNames {
		def := g.Schema.Types[typeName]
		if def == nil {
			continue
		}

		operations := g.generateOperationsForType(def)
		if operations != "" {
			filename := strings.ToLower(typeName) + "s.graphql"
			if typeName == "Query" {
				filename = "queries.graphql"
			} else if typeName == "Mutation" {
				filename = "mutations.graphql"
			}
			result[filename] = operations
		}
	}

	return result, nil
}

func (g *Generator) generateOperationsForType(def *ast.Definition) string {
	if def == nil || len(def.Fields) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Auto-generated %s operations\n\n", def.Name))

	// Sort fields for consistent output
	fields := make([]*ast.FieldDefinition, len(def.Fields))
	copy(fields, def.Fields)
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})

	for _, field := range fields {
		// Skip if doesn't match include pattern
		if !g.shouldInclude(field.Name) {
			continue
		}

		operation := g.generateOperation(field, def.Name)
		if operation != "" {
			sb.WriteString(operation)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func (g *Generator) shouldInclude(fieldName string) bool {
	// Check exclude patterns first
	for _, pattern := range g.Exclude {
		if matched, _ := doublestar.Match(pattern, fieldName); matched {
			return false
		}
	}

	// If no include patterns, include everything
	if len(g.Include) == 0 {
		return true
	}

	// Check include patterns
	for _, pattern := range g.Include {
		if matched, _ := doublestar.Match(pattern, fieldName); matched {
			return true
		}
	}

	return false
}

func (g *Generator) generateOperation(field *ast.FieldDefinition, typeName string) string {
	var sb strings.Builder

	// Generate operation name
	opType := "query"
	if typeName == "Mutation" {
		opType = "mutation"
	}
	opName := toOperationName(field.Name, opType)

	// Collect arguments
	var args []string
	var params []string
	for _, arg := range field.Arguments {
		argType := arg.Type.String()
		args = append(args, fmt.Sprintf("$%s: %s", arg.Name, argType))
		params = append(params, fmt.Sprintf("%s: $%s", arg.Name, arg.Name))
	}

	// Generate operation signature
	if len(args) > 0 {
		sb.WriteString(fmt.Sprintf("%s %s(%s) {\n", opType, opName, strings.Join(args, ", ")))
	} else {
		sb.WriteString(fmt.Sprintf("%s %s {\n", opType, opName))
	}

	// Generate field call with parameters
	if len(params) > 0 {
		sb.WriteString(fmt.Sprintf("  %s(%s)", field.Name, strings.Join(params, ", ")))
	} else {
		sb.WriteString(fmt.Sprintf("  %s", field.Name))
	}

	// Generate selection set
	selectionSet := g.generateSelectionSet(field.Type, 1)
	if selectionSet != "" {
		sb.WriteString(" {\n")
		sb.WriteString(selectionSet)
		sb.WriteString("  }\n")
	} else {
		sb.WriteString("\n")
	}

	sb.WriteString("}\n")
	return sb.String()
}

func (g *Generator) generateSelectionSet(typeRef *ast.Type, depth int) string {
	if depth > g.MaxDepth {
		return ""
	}

	// Get the underlying type name
	typeName := getTypeName(typeRef)
	if typeName == "" {
		return ""
	}

	// Get type definition
	def := g.Schema.Types[typeName]
	if def == nil {
		return ""
	}

	// Only generate selection sets for OBJECT and INTERFACE types
	if def.Kind != ast.Object && def.Kind != ast.Interface {
		return ""
	}

	var sb strings.Builder
	indent := strings.Repeat("    ", depth)

	// Sort fields for consistent output
	fields := make([]*ast.FieldDefinition, len(def.Fields))
	copy(fields, def.Fields)
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Name < fields[j].Name
	})

	for _, field := range fields {
		// Skip fields that start with double underscore (introspection)
		if strings.HasPrefix(field.Name, "__") {
			continue
		}

		fieldTypeName := getTypeName(field.Type)
		fieldDef := g.Schema.Types[fieldTypeName]

		// Check if this field needs a selection set
		needsSelection := fieldDef != nil && (fieldDef.Kind == ast.Object || fieldDef.Kind == ast.Interface)

		if needsSelection {
			if depth < g.MaxDepth {
				nestedSelection := g.generateSelectionSet(field.Type, depth+1)
				if nestedSelection != "" {
					sb.WriteString(fmt.Sprintf("%s%s {\n", indent, field.Name))
					sb.WriteString(nestedSelection)
					sb.WriteString(fmt.Sprintf("%s}\n", indent))
				}
			}
			// Skip complex fields at max depth
		} else {
			// Scalar or enum field
			sb.WriteString(fmt.Sprintf("%s%s\n", indent, field.Name))
		}
	}

	return sb.String()
}

func getTypeName(t *ast.Type) string {
	if t == nil {
		return ""
	}
	if t.NamedType != "" {
		return t.NamedType
	}
	if t.Elem != nil {
		return getTypeName(t.Elem)
	}
	return ""
}

func toOperationName(fieldName, opType string) string {
	// Convert field name to operation name
	// e.g., "user" -> "GetUser" for queries, "createUser" -> "CreateUser" for mutations
	name := strings.ToUpper(fieldName[:1]) + fieldName[1:]
	if opType == "query" {
		// Add "Get" prefix if not already present
		if !strings.HasPrefix(strings.ToLower(name), "get") &&
			!strings.HasPrefix(strings.ToLower(name), "list") &&
			!strings.HasPrefix(strings.ToLower(name), "find") &&
			!strings.HasPrefix(strings.ToLower(name), "search") {
			name = "Get" + name
		}
	}
	return name
}

// ParsePatterns parses comma-separated glob patterns.
func ParsePatterns(pattern string) []string {
	if pattern == "" {
		return nil
	}
	parts := strings.Split(pattern, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// WriteFiles writes the generated files to the output directory.
func WriteFiles(files map[string]string, outputDir string) error {
	for filename, content := range files {
		path := filepath.Join(outputDir, filename)
		if err := writeFile(path, content); err != nil {
			return err
		}
	}
	return nil
}

func writeFile(path, content string) error {
	// This is a stub - actual implementation is in the command
	return nil
}
