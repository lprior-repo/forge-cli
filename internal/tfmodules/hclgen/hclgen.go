// Package hclgen provides HCL generation utilities for tfmodules.
// This implements type-safe Terraform HCL generation following functional programming principles.
package hclgen

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// ToHCL converts a module struct to HCL format.
// PURE: Same input always produces same output.
func ToHCL(localName, source, version string, v interface{}) (string, error) {
	var parts []string

	// Module header
	parts = append(parts, fmt.Sprintf("module \"%s\" {", localName))
	parts = append(parts, fmt.Sprintf("  source  = \"%s\"", source))
	if version != "" {
		parts = append(parts, fmt.Sprintf("  version = \"%s\"", version))
	}
	parts = append(parts, "")

	// Convert struct fields to HCL
	hclFields, err := structToHCL(v, "  ")
	if err != nil {
		return "", err
	}

	parts = append(parts, hclFields...)
	parts = append(parts, "}")

	return strings.Join(parts, "\n"), nil
}

// structToHCL converts a struct to HCL field definitions.
// PURE: Deterministic conversion based on struct tags.
func structToHCL(v interface{}, indent string) ([]string, error) {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	// Dereference pointer if necessary
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected struct, got %v", val.Kind())
	}

	var lines []string
	var fieldGroups []fieldGroup

	// Group fields for organized output
	for i := 0; i < val.NumField(); i++ {
		field := typ.Field(i)
		fieldVal := val.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Skip special fields
		if field.Name == "Source" || field.Name == "Version" || field.Name == "Region" {
			continue
		}

		// Get HCL tag
		hclTag := field.Tag.Get("hcl")
		if hclTag == "" || hclTag == "-" {
			continue
		}

		// Parse HCL tag (format: "name,attr" or "name,block")
		parts := strings.Split(hclTag, ",")
		if len(parts) == 0 {
			continue
		}

		attrName := parts[0]
		isBlock := len(parts) > 1 && parts[1] == "block"

		// Skip nil pointers (not set)
		if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
			continue
		}

		// Skip empty slices and maps
		if (fieldVal.Kind() == reflect.Slice || fieldVal.Kind() == reflect.Map) && fieldVal.Len() == 0 {
			continue
		}

		fg := fieldGroup{
			name:    attrName,
			isBlock: isBlock,
			value:   fieldVal,
		}

		fieldGroups = append(fieldGroups, fg)
	}

	// Sort fields for consistent output
	sort.Slice(fieldGroups, func(i, j int) bool {
		return fieldGroups[i].name < fieldGroups[j].name
	})

	// Convert field groups to HCL
	for _, fg := range fieldGroups {
		hcl, err := valueToHCL(fg.name, fg.value, indent, fg.isBlock)
		if err != nil {
			return nil, err
		}
		if hcl != "" {
			lines = append(lines, hcl)
		}
	}

	return lines, nil
}

// fieldGroup represents a struct field for HCL generation.
type fieldGroup struct {
	name    string
	isBlock bool
	value   reflect.Value
}

// valueToHCL converts a reflect.Value to HCL representation.
// PURE: Deterministic conversion based on value type.
func valueToHCL(name string, val reflect.Value, indent string, isBlock bool) (string, error) {
	// Dereference pointer
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return "", nil
		}
		val = val.Elem()
	}

	// Handle interface{} by extracting the concrete value
	if val.Kind() == reflect.Interface {
		if val.IsNil() {
			return "", nil
		}
		val = val.Elem()
	}

	switch val.Kind() {
	case reflect.String:
		s := val.String()
		// Check if it's a Terraform reference (starts with ${)
		if strings.HasPrefix(s, "${") {
			return fmt.Sprintf("%s%s = %s", indent, name, s), nil
		}
		return fmt.Sprintf("%s%s = %q", indent, name, s), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return fmt.Sprintf("%s%s = %d", indent, name, val.Int()), nil

	case reflect.Bool:
		return fmt.Sprintf("%s%s = %t", indent, name, val.Bool()), nil

	case reflect.Map:
		return mapToHCL(name, val, indent, isBlock)

	case reflect.Slice:
		return sliceToHCL(name, val, indent, isBlock)

	case reflect.Struct:
		return structBlockToHCL(name, val, indent)

	default:
		return "", fmt.Errorf("unsupported type for %s: %v", name, val.Kind())
	}
}

// mapToHCL converts a map to HCL format.
// PURE: Deterministic map conversion.
func mapToHCL(name string, val reflect.Value, indent string, isBlock bool) (string, error) {
	if val.Len() == 0 {
		return "", nil
	}

	var lines []string

	if isBlock {
		// Block format: name { key = value }
		lines = append(lines, fmt.Sprintf("%s%s {", indent, name))
		nextIndent := indent + "  "

		// Sort keys for deterministic output
		keys := val.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, key := range keys {
			v := val.MapIndex(key)
			hcl, err := valueToHCL(key.String(), v, nextIndent, false)
			if err != nil {
				return "", err
			}
			lines = append(lines, hcl)
		}

		lines = append(lines, fmt.Sprintf("%s}", indent))
	} else {
		// Attribute format: name = { key = value }
		lines = append(lines, fmt.Sprintf("%s%s = {", indent, name))
		nextIndent := indent + "  "

		keys := val.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, key := range keys {
			v := val.MapIndex(key)

			// Handle nested maps/interfaces
			if v.Kind() == reflect.Interface {
				v = v.Elem()
			}

			hcl, err := valueToHCL(key.String(), v, nextIndent, false)
			if err != nil {
				return "", err
			}
			lines = append(lines, hcl)
		}

		lines = append(lines, fmt.Sprintf("%s}", indent))
	}

	return strings.Join(lines, "\n"), nil
}

// sliceToHCL converts a slice to HCL format.
// PURE: Deterministic slice conversion.
func sliceToHCL(name string, val reflect.Value, indent string, isBlock bool) (string, error) {
	if val.Len() == 0 {
		return "", nil
	}

	// Check element type
	elemType := val.Type().Elem()

	if isBlock {
		// Repeated blocks: name { ... } name { ... }
		var lines []string
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			if elem.Kind() == reflect.Struct {
				hcl, err := structBlockToHCL(name, elem, indent)
				if err != nil {
					return "", err
				}
				lines = append(lines, hcl)
			}
		}
		return strings.Join(lines, "\n"), nil
	}

	// Array format: name = [...]
	if elemType.Kind() == reflect.String {
		var items []string
		for i := 0; i < val.Len(); i++ {
			s := val.Index(i).String()
			// Check if it's a Terraform reference
			if strings.HasPrefix(s, "${") {
				items = append(items, s)
			} else {
				items = append(items, fmt.Sprintf("%q", s))
			}
		}
		return fmt.Sprintf("%s%s = [%s]", indent, name, strings.Join(items, ", ")), nil
	}

	if elemType.Kind() == reflect.Struct || elemType.Kind() == reflect.Map {
		// Array of objects: name = [{...}, {...}]
		var items []string
		for i := 0; i < val.Len(); i++ {
			elem := val.Index(i)
			hcl, err := objectToHCL(elem, indent+"  ")
			if err != nil {
				return "", err
			}
			items = append(items, hcl)
		}
		return fmt.Sprintf("%s%s = [\n%s\n%s]", indent, name, strings.Join(items, ",\n"), indent), nil
	}

	return "", fmt.Errorf("unsupported slice element type: %v", elemType.Kind())
}

// structBlockToHCL converts a struct to a block format.
// PURE: Deterministic struct block conversion.
func structBlockToHCL(name string, val reflect.Value, indent string) (string, error) {
	var lines []string

	lines = append(lines, fmt.Sprintf("%s%s {", indent, name))
	nextIndent := indent + "  "

	fields, err := structToHCL(val.Interface(), nextIndent)
	if err != nil {
		return "", err
	}

	lines = append(lines, fields...)
	lines = append(lines, fmt.Sprintf("%s}", indent))

	return strings.Join(lines, "\n"), nil
}

// objectToHCL converts an object (struct or map) to HCL object syntax.
// PURE: Deterministic object conversion.
func objectToHCL(val reflect.Value, indent string) (string, error) {
	var lines []string

	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	lines = append(lines, indent+"{")
	nextIndent := indent + "  "

	if val.Kind() == reflect.Struct {
		fields, err := structToHCL(val.Interface(), nextIndent)
		if err != nil {
			return "", err
		}
		lines = append(lines, fields...)
	} else if val.Kind() == reflect.Map {
		keys := val.MapKeys()
		sort.Slice(keys, func(i, j int) bool {
			return keys[i].String() < keys[j].String()
		})

		for _, key := range keys {
			v := val.MapIndex(key)
			hcl, err := valueToHCL(key.String(), v, nextIndent, false)
			if err != nil {
				return "", err
			}
			lines = append(lines, hcl)
		}
	}

	lines = append(lines, indent+"}")

	return strings.Join(lines, "\n"), nil
}
