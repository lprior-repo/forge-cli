// Package hclgen provides HCL generation utilities using hashicorp/hcl/v2/hclwrite.
// This implements type-safe Terraform HCL generation using the official HashiCorp library.
package hclgen

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

// ToHCLWrite converts a module struct to HCL format using hclwrite.
// PURE: Same input always produces same output.
func ToHCLWrite(localName, source, version string, v interface{}) (string, error) {
	f := hclwrite.NewEmptyFile()
	rootBody := f.Body()

	// Create module block
	moduleBlock := rootBody.AppendNewBlock("module", []string{localName})
	moduleBody := moduleBlock.Body()

	// Add source and version
	moduleBody.SetAttributeValue("source", cty.StringVal(source))
	if version != "" {
		moduleBody.SetAttributeValue("version", cty.StringVal(version))
	}

	// Convert struct fields to HCL attributes
	if err := structToHCLWrite(moduleBody, v); err != nil {
		return "", err
	}

	return string(f.Bytes()), nil
}

// structToHCLWrite converts a struct to HCL attributes and blocks using hclwrite.
// PURE: Deterministic conversion based on struct tags.
func structToHCLWrite(body *hclwrite.Body, v interface{}) error {
	val := reflect.ValueOf(v)
	typ := reflect.TypeOf(v)

	// Dereference pointer if necessary
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
		typ = typ.Elem()
	}

	if val.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %v", val.Kind())
	}

	// Collect and sort fields for deterministic output
	var attrs []fieldInfo
	var blocks []fieldInfo

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

		info := fieldInfo{
			name:    attrName,
			isBlock: isBlock,
			value:   fieldVal,
		}

		if isBlock {
			blocks = append(blocks, info)
		} else {
			attrs = append(attrs, info)
		}
	}

	// Sort fields for deterministic output
	sort.Slice(attrs, func(i, j int) bool {
		return attrs[i].name < attrs[j].name
	})
	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].name < blocks[j].name
	})

	// Write attributes first
	for _, field := range attrs {
		if err := setAttribute(body, field.name, field.value); err != nil {
			return fmt.Errorf("attribute %s: %w", field.name, err)
		}
	}

	// Write blocks
	for _, field := range blocks {
		if err := writeBlock(body, field.name, field.value); err != nil {
			return fmt.Errorf("block %s: %w", field.name, err)
		}
	}

	return nil
}

// setAttribute sets an attribute in the HCL body, handling Terraform references.
// PURE: Deterministic attribute setting.
func setAttribute(body *hclwrite.Body, name string, v reflect.Value) error {
	// Dereference pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	// Handle interface{} by extracting concrete value
	if v.Kind() == reflect.Interface {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		// Detect Terraform references (${...} or unquoted references)
		if isTerraformReference(s) {
			// Use SetAttributeRaw for Terraform expressions
			tokens := tokenizeReference(s)
			body.SetAttributeRaw(name, tokens)
		} else {
			body.SetAttributeValue(name, cty.StringVal(s))
		}
		return nil

	case reflect.Bool:
		body.SetAttributeValue(name, cty.BoolVal(v.Bool()))
		return nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		body.SetAttributeValue(name, cty.NumberIntVal(v.Int()))
		return nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		body.SetAttributeValue(name, cty.NumberUIntVal(v.Uint()))
		return nil

	case reflect.Float32, reflect.Float64:
		body.SetAttributeValue(name, cty.NumberFloatVal(v.Float()))
		return nil

	case reflect.Slice, reflect.Array:
		return setSliceAttribute(body, name, v)

	case reflect.Map:
		return setMapAttribute(body, name, v)

	case reflect.Struct:
		// Structs as attributes should be converted to objects
		ctyVal, err := goValueToCty(v)
		if err != nil {
			return err
		}
		body.SetAttributeValue(name, ctyVal)
		return nil

	default:
		return fmt.Errorf("unsupported type for attribute %s: %v", name, v.Kind())
	}
}

// setSliceAttribute sets a slice/array attribute.
// PURE: Deterministic slice conversion.
func setSliceAttribute(body *hclwrite.Body, name string, v reflect.Value) error {
	if v.Len() == 0 {
		return nil
	}

	// Check if any element is a Terraform reference
	hasReferences := false
	for i := 0; i < v.Len(); i++ {
		elem := v.Index(i)
		if elem.Kind() == reflect.String && isTerraformReference(elem.String()) {
			hasReferences = true
			break
		}
	}

	if hasReferences {
		// Build raw token list for array with references
		tokens := hclwrite.Tokens{{Type: 91, Bytes: []byte("[")}} // [
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				tokens = append(tokens, &hclwrite.Token{Type: 44, Bytes: []byte(",")})  // ,
				tokens = append(tokens, &hclwrite.Token{Type: 10, Bytes: []byte(" ")}) // space
			}
			elem := v.Index(i)
			if elem.Kind() == reflect.String {
				s := elem.String()
				if isTerraformReference(s) {
					// Add reference without quotes
					tokens = append(tokens, tokenizeReference(s)...)
				} else {
					// Add quoted string
					tokens = append(tokens, &hclwrite.Token{Type: 34, Bytes: []byte(`"` + s + `"`)})
				}
			}
		}
		tokens = append(tokens, &hclwrite.Token{Type: 93, Bytes: []byte("]")}) // ]
		body.SetAttributeRaw(name, tokens)
		return nil
	}

	// No references - use standard cty conversion
	ctyVal, err := goValueToCty(v)
	if err != nil {
		return err
	}
	body.SetAttributeValue(name, ctyVal)
	return nil
}

// setMapAttribute sets a map attribute.
// PURE: Deterministic map conversion.
func setMapAttribute(body *hclwrite.Body, name string, v reflect.Value) error {
	if v.Len() == 0 {
		return nil
	}

	// Check if any value is a Terraform reference
	hasReferences := false
	iter := v.MapRange()
	for iter.Next() {
		val := iter.Value()
		if val.Kind() == reflect.String && isTerraformReference(val.String()) {
			hasReferences = true
			break
		} else if val.Kind() == reflect.Interface {
			elem := val.Elem()
			if elem.Kind() == reflect.String && isTerraformReference(elem.String()) {
				hasReferences = true
				break
			}
		}
	}

	if hasReferences {
		// Build raw token map with sorted keys
		keys := make([]string, 0, v.Len())
		iter = v.MapRange()
		for iter.Next() {
			keys = append(keys, iter.Key().String())
		}
		sort.Strings(keys)

		tokens := hclwrite.Tokens{{Type: 123, Bytes: []byte("{")}} // {
		tokens = append(tokens, &hclwrite.Token{Type: 10, Bytes: []byte("\n")})

		for _, key := range keys {
			mapVal := v.MapIndex(reflect.ValueOf(key))

			// Add indentation
			tokens = append(tokens, &hclwrite.Token{Type: 10, Bytes: []byte("    ")})
			// Add key
			tokens = append(tokens, &hclwrite.Token{Type: 1, Bytes: []byte(key)})
			// Add =
			tokens = append(tokens, &hclwrite.Token{Type: 10, Bytes: []byte(" ")})
			tokens = append(tokens, &hclwrite.Token{Type: 61, Bytes: []byte("=")})
			tokens = append(tokens, &hclwrite.Token{Type: 10, Bytes: []byte(" ")})

			// Add value
			if mapVal.Kind() == reflect.Interface {
				mapVal = mapVal.Elem()
			}

			if mapVal.Kind() == reflect.String {
				s := mapVal.String()
				if isTerraformReference(s) {
					tokens = append(tokens, tokenizeReference(s)...)
				} else {
					tokens = append(tokens, &hclwrite.Token{Type: 34, Bytes: []byte(`"` + s + `"`)})
				}
			} else {
				// Use cty for non-string values
				ctyVal, err := goValueToCty(mapVal)
				if err != nil {
					return err
				}
				valTokens := hclwrite.TokensForValue(ctyVal)
				tokens = append(tokens, valTokens...)
			}

			tokens = append(tokens, &hclwrite.Token{Type: 10, Bytes: []byte("\n")})
		}

		tokens = append(tokens, &hclwrite.Token{Type: 10, Bytes: []byte("  ")})
		tokens = append(tokens, &hclwrite.Token{Type: 125, Bytes: []byte("}")}) // }
		body.SetAttributeRaw(name, tokens)
		return nil
	}

	// No references - use standard cty conversion
	ctyVal, err := goValueToCty(v)
	if err != nil {
		return err
	}
	body.SetAttributeValue(name, ctyVal)
	return nil
}

// isTerraformReference checks if a string is a Terraform reference.
// PURE: Deterministic check.
func isTerraformReference(s string) bool {
	// Check for interpolation syntax: ${...}
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		return true
	}
	// Check for direct references: var.name, module.name.output, etc.
	if strings.HasPrefix(s, "var.") ||
		strings.HasPrefix(s, "module.") ||
		strings.HasPrefix(s, "data.") ||
		strings.HasPrefix(s, "resource.") ||
		strings.HasPrefix(s, "local.") {
		return true
	}
	return false
}

// tokenizeReference converts a Terraform reference string to hclwrite tokens.
// PURE: Deterministic tokenization.
func tokenizeReference(s string) hclwrite.Tokens {
	// For ${...} syntax, strip the delimiters
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		s = s[2 : len(s)-1]
	}

	// Return as identifier token
	return hclwrite.Tokens{
		{Type: 1, Bytes: []byte(s)}, // IDENT token type
	}
}

// goValueToCty converts a Go reflect.Value to a cty.Value.
// PURE: Deterministic conversion.
func goValueToCty(v reflect.Value) (cty.Value, error) {
	// Dereference pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return cty.NullVal(cty.DynamicPseudoType), nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		return cty.StringVal(v.String()), nil

	case reflect.Bool:
		return cty.BoolVal(v.Bool()), nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return cty.NumberIntVal(v.Int()), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return cty.NumberUIntVal(v.Uint()), nil

	case reflect.Float32, reflect.Float64:
		return cty.NumberFloatVal(v.Float()), nil

	case reflect.Slice, reflect.Array:
		if v.Len() == 0 {
			return cty.ListValEmpty(cty.DynamicPseudoType), nil
		}

		vals := make([]cty.Value, v.Len())
		for i := 0; i < v.Len(); i++ {
			elemVal, err := goValueToCty(v.Index(i))
			if err != nil {
				return cty.NilVal, fmt.Errorf("slice element %d: %w", i, err)
			}
			vals[i] = elemVal
		}
		return cty.ListVal(vals), nil

	case reflect.Map:
		if v.Len() == 0 {
			return cty.MapValEmpty(cty.DynamicPseudoType), nil
		}

		vals := make(map[string]cty.Value)
		iter := v.MapRange()
		for iter.Next() {
			key := iter.Key()
			val := iter.Value()

			keyStr, ok := key.Interface().(string)
			if !ok {
				return cty.NilVal, fmt.Errorf("map key must be string, got %v", key.Kind())
			}

			ctyVal, err := goValueToCty(val)
			if err != nil {
				return cty.NilVal, fmt.Errorf("map value for key %s: %w", keyStr, err)
			}
			vals[keyStr] = ctyVal
		}
		return cty.MapVal(vals), nil

	case reflect.Struct:
		// Convert struct to cty.Object
		typ := v.Type()
		objVals := make(map[string]cty.Value)
		objTypes := make(map[string]cty.Type)

		for i := 0; i < v.NumField(); i++ {
			field := typ.Field(i)
			if !field.IsExported() {
				continue
			}

			hclTag := field.Tag.Get("hcl")
			if hclTag == "" || hclTag == "-" {
				continue
			}

			parts := strings.Split(hclTag, ",")
			attrName := parts[0]

			fieldVal := v.Field(i)
			ctyVal, err := goValueToCty(fieldVal)
			if err != nil {
				return cty.NilVal, fmt.Errorf("struct field %s: %w", field.Name, err)
			}

			objVals[attrName] = ctyVal
			objTypes[attrName] = ctyVal.Type()
		}

		if len(objVals) == 0 {
			return cty.EmptyObjectVal, nil
		}

		return cty.ObjectVal(objVals), nil

	case reflect.Interface:
		if v.IsNil() {
			return cty.NullVal(cty.DynamicPseudoType), nil
		}
		return goValueToCty(v.Elem())

	default:
		return cty.NilVal, fmt.Errorf("unsupported type: %v", v.Kind())
	}
}

// writeBlock writes a block to the HCL body.
// PURE: Side-effect free (modifies body parameter).
func writeBlock(body *hclwrite.Body, blockType string, v reflect.Value) error {
	// Dereference pointers
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		// Single block
		block := body.AppendNewBlock(blockType, nil)
		return structToHCLWrite(block.Body(), v.Interface())

	case reflect.Slice, reflect.Array:
		// Multiple blocks
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Struct {
				block := body.AppendNewBlock(blockType, nil)
				if err := structToHCLWrite(block.Body(), elem.Interface()); err != nil {
					return fmt.Errorf("block %d: %w", i, err)
				}
			} else if elem.Kind() == reflect.Ptr && !elem.IsNil() {
				block := body.AppendNewBlock(blockType, nil)
				if err := structToHCLWrite(block.Body(), elem.Elem().Interface()); err != nil {
					return fmt.Errorf("block %d: %w", i, err)
				}
			}
		}
		return nil

	case reflect.Map:
		// Map blocks (key becomes block label) - sort for determinism
		keys := make([]string, 0, v.Len())
		iter := v.MapRange()
		for iter.Next() {
			keyStr, ok := iter.Key().Interface().(string)
			if !ok {
				return fmt.Errorf("block map key must be string, got %v", iter.Key().Kind())
			}
			keys = append(keys, keyStr)
		}
		sort.Strings(keys)

		for _, keyStr := range keys {
			val := v.MapIndex(reflect.ValueOf(keyStr))

			if val.Kind() == reflect.Struct {
				block := body.AppendNewBlock(blockType, []string{keyStr})
				if err := structToHCLWrite(block.Body(), val.Interface()); err != nil {
					return fmt.Errorf("block %s: %w", keyStr, err)
				}
			} else if val.Kind() == reflect.Ptr && !val.IsNil() {
				block := body.AppendNewBlock(blockType, []string{keyStr})
				if err := structToHCLWrite(block.Body(), val.Elem().Interface()); err != nil {
					return fmt.Errorf("block %s: %w", keyStr, err)
				}
			}
		}
		return nil

	default:
		return fmt.Errorf("block must be struct, slice, or map, got %v", v.Kind())
	}
}

// fieldInfo holds information about a struct field for HCL generation.
type fieldInfo struct {
	name    string
	isBlock bool
	value   reflect.Value
}
