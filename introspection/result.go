package introspection

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Result contains the introspection result with conversion methods.
type Result struct {
	Schema   Schema `json:"schema"`
	Endpoint string `json:"endpoint,omitempty"`
}

// ToJSON converts the introspection result to JSON.
func (r *Result) ToJSON() ([]byte, error) {
	return json.MarshalIndent(r, "", "  ")
}

// ToSDL converts the introspection result to SDL (Schema Definition Language).
func (r *Result) ToSDL() (string, error) {
	var sb strings.Builder

	// Write schema definition if we have custom operation types
	if r.Schema.QueryType != nil || r.Schema.MutationType != nil || r.Schema.SubscriptionType != nil {
		sb.WriteString("schema {\n")
		if r.Schema.QueryType != nil {
			fmt.Fprintf(&sb, "  query: %s\n", r.Schema.QueryType.Name)
		}
		if r.Schema.MutationType != nil {
			fmt.Fprintf(&sb, "  mutation: %s\n", r.Schema.MutationType.Name)
		}
		if r.Schema.SubscriptionType != nil {
			fmt.Fprintf(&sb, "  subscription: %s\n", r.Schema.SubscriptionType.Name)
		}
		sb.WriteString("}\n\n")
	}

	// Sort types for consistent output
	types := make([]FullType, len(r.Schema.Types))
	copy(types, r.Schema.Types)
	sort.Slice(types, func(i, j int) bool {
		return types[i].Name < types[j].Name
	})

	// Write types
	for _, t := range types {
		// Skip built-in types (starting with __)
		if strings.HasPrefix(t.Name, "__") {
			continue
		}

		sdl, err := typeToSDL(t)
		if err != nil {
			return "", err
		}
		if sdl != "" {
			sb.WriteString(sdl)
			sb.WriteString("\n")
		}
	}

	// Write directives (excluding built-ins)
	for _, d := range r.Schema.Directives {
		if isBuiltInDirective(d.Name) {
			continue
		}
		sdl := directiveToSDL(d)
		sb.WriteString(sdl)
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

func typeToSDL(t FullType) (string, error) {
	var sb strings.Builder

	// Write description
	if t.Description != "" {
		sb.WriteString(formatDescription(t.Description, ""))
	}

	switch t.Kind {
	case KindScalar:
		// Skip built-in scalars
		if isBuiltInScalar(t.Name) {
			return "", nil
		}
		fmt.Fprintf(&sb, "scalar %s\n", t.Name)

	case KindObject:
		fmt.Fprintf(&sb, "type %s", t.Name)
		if len(t.Interfaces) > 0 {
			ifaces := make([]string, len(t.Interfaces))
			for i, iface := range t.Interfaces {
				ifaces[i] = iface.Name
			}
			fmt.Fprintf(&sb, " implements %s", strings.Join(ifaces, " & "))
		}
		sb.WriteString(" {\n")
		for _, f := range t.Fields {
			sb.WriteString(fieldToSDL(f, "  "))
		}
		sb.WriteString("}\n")

	case KindInterface:
		fmt.Fprintf(&sb, "interface %s {\n", t.Name)
		for _, f := range t.Fields {
			sb.WriteString(fieldToSDL(f, "  "))
		}
		sb.WriteString("}\n")

	case KindUnion:
		types := make([]string, len(t.PossibleTypes))
		for i, pt := range t.PossibleTypes {
			types[i] = pt.Name
		}
		fmt.Fprintf(&sb, "union %s = %s\n", t.Name, strings.Join(types, " | "))

	case KindEnum:
		fmt.Fprintf(&sb, "enum %s {\n", t.Name)
		for _, ev := range t.EnumValues {
			if ev.Description != "" {
				sb.WriteString(formatDescription(ev.Description, "  "))
			}
			fmt.Fprintf(&sb, "  %s", ev.Name)
			if ev.IsDeprecated {
				if ev.DeprecationReason != "" {
					fmt.Fprintf(&sb, " @deprecated(reason: %q)", ev.DeprecationReason)
				} else {
					sb.WriteString(" @deprecated")
				}
			}
			sb.WriteString("\n")
		}
		sb.WriteString("}\n")

	case KindInputObject:
		fmt.Fprintf(&sb, "input %s {\n", t.Name)
		for _, f := range t.InputFields {
			sb.WriteString(inputValueToSDL(f, "  ", false))
		}
		sb.WriteString("}\n")
	}

	return sb.String(), nil
}

func fieldToSDL(f Field, indent string) string {
	var sb strings.Builder

	if f.Description != "" {
		sb.WriteString(formatDescription(f.Description, indent))
	}

	sb.WriteString(indent)
	sb.WriteString(f.Name)

	// Arguments
	if len(f.Args) > 0 {
		sb.WriteString("(")
		args := make([]string, len(f.Args))
		for i, arg := range f.Args {
			args[i] = strings.TrimSpace(inputValueToSDL(arg, "", true))
		}
		sb.WriteString(strings.Join(args, ", "))
		sb.WriteString(")")
	}

	sb.WriteString(": ")
	sb.WriteString(typeRefToSDL(f.Type))

	if f.IsDeprecated {
		if f.DeprecationReason != "" {
			fmt.Fprintf(&sb, " @deprecated(reason: %q)", f.DeprecationReason)
		} else {
			sb.WriteString(" @deprecated")
		}
	}

	sb.WriteString("\n")
	return sb.String()
}

func inputValueToSDL(iv InputValue, indent string, inline bool) string {
	var sb strings.Builder

	if !inline && iv.Description != "" {
		sb.WriteString(formatDescription(iv.Description, indent))
	}

	sb.WriteString(indent)
	sb.WriteString(iv.Name)
	sb.WriteString(": ")
	sb.WriteString(typeRefToSDL(iv.Type))

	if iv.DefaultValue != nil {
		fmt.Fprintf(&sb, " = %s", *iv.DefaultValue)
	}

	if !inline {
		sb.WriteString("\n")
	}

	return sb.String()
}

func typeRefToSDL(tr TypeRef) string {
	switch tr.Kind {
	case KindNonNull:
		if tr.OfType != nil {
			return typeRefToSDL(*tr.OfType) + "!"
		}
		return "!"
	case KindList:
		if tr.OfType != nil {
			return "[" + typeRefToSDL(*tr.OfType) + "]"
		}
		return "[]"
	default:
		return tr.Name
	}
}

func directiveToSDL(d Directive) string {
	var sb strings.Builder

	if d.Description != "" {
		sb.WriteString(formatDescription(d.Description, ""))
	}

	fmt.Fprintf(&sb, "directive @%s", d.Name)

	if len(d.Args) > 0 {
		sb.WriteString("(")
		args := make([]string, len(d.Args))
		for i, arg := range d.Args {
			args[i] = strings.TrimSpace(inputValueToSDL(arg, "", true))
		}
		sb.WriteString(strings.Join(args, ", "))
		sb.WriteString(")")
	}

	if len(d.Locations) > 0 {
		sb.WriteString(" on ")
		sb.WriteString(strings.Join(d.Locations, " | "))
	}

	sb.WriteString("\n")
	return sb.String()
}

func formatDescription(desc, indent string) string {
	if desc == "" {
		return ""
	}
	// Use block string for multi-line descriptions
	if strings.Contains(desc, "\n") {
		return fmt.Sprintf("%s\"\"\"\n%s%s\n%s\"\"\"\n", indent, indent, desc, indent)
	}
	return fmt.Sprintf("%s\"%s\"\n", indent, escapeString(desc))
}

func escapeString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

func isBuiltInScalar(name string) bool {
	builtIns := map[string]bool{
		"String":  true,
		"Int":     true,
		"Float":   true,
		"Boolean": true,
		"ID":      true,
	}
	return builtIns[name]
}

func isBuiltInDirective(name string) bool {
	builtIns := map[string]bool{
		"skip":        true,
		"include":     true,
		"deprecated":  true,
		"specifiedBy": true,
	}
	return builtIns[name]
}
