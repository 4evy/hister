package searchschema

import "testing"

func TestDefinitionsAreInternallyConsistent(t *testing.T) {
	seenFields := make(map[string]bool)
	for _, field := range Fields() {
		if seenFields[field.Name] {
			t.Fatalf("duplicate field %q", field.Name)
		}
		seenFields[field.Name] = true
		if field.IndexField == "" {
			t.Fatalf("field %q has no index field", field.Name)
		}
		if field.ValueSet != "" && len(Values(field.ValueSet)) == 0 {
			t.Fatalf("field %q references empty value set %q", field.Name, field.ValueSet)
		}
	}

	seenFacets := make(map[string]bool)
	for _, facet := range Facets() {
		if seenFacets[facet.Name] {
			t.Fatalf("duplicate facet %q", facet.Name)
		}
		seenFacets[facet.Name] = true
		field, ok := Field(facet.QueryField)
		if !ok {
			t.Fatalf("facet %q references unknown query field %q", facet.Name, facet.QueryField)
		}
		if field.ValueSet != "" && len(Values(field.ValueSet)) == 0 {
			t.Fatalf("facet %q references field with empty value set %q", facet.Name, field.ValueSet)
		}
	}

	defaultSorts := 0
	seenSorts := make(map[string]bool)
	for _, sort := range CapabilitiesDefinition().Sort.Options {
		if seenSorts[sort.Value] {
			t.Fatalf("duplicate sort %q", sort.Value)
		}
		seenSorts[sort.Value] = true
		if len(sort.Fields) == 0 {
			t.Fatalf("sort %q has no Bleve fields", sort.Value)
		}
		if sort.Default {
			defaultSorts++
		}
	}
	if defaultSorts != 1 {
		t.Fatalf("default sort count = %d, want 1", defaultSorts)
	}
}

func TestCapabilitiesContainOnlyVisibleFields(t *testing.T) {
	capabilities := CapabilitiesDefinition()
	if capabilities.Version != Version {
		t.Fatalf("capability version = %d, want %d", capabilities.Version, Version)
	}
	for _, field := range capabilities.Fields {
		if !field.Visible {
			t.Fatalf("hidden field %q exposed in capabilities", field.Name)
		}
	}
	fieldsByName := make(map[string]FieldDefinition, len(capabilities.Fields))
	for _, field := range capabilities.Fields {
		fieldsByName[field.Name] = field
	}
	for _, facet := range capabilities.Facets {
		field := fieldsByName[facet.QueryField]
		if field.Facet != facet.Name {
			t.Fatalf("public field %q references facet %q, want %q", field.Name, field.Facet, facet.Name)
		}
		if facet.ValueSet != field.ValueSet {
			t.Fatalf("public facet %q value set = %q, want %q", facet.Name, facet.ValueSet, field.ValueSet)
		}
	}
	if _, ok := Field("user_id"); !ok {
		t.Fatal("internal user_id field is missing")
	}
	for _, field := range capabilities.Fields {
		if field.Name == "user_id" {
			t.Fatal("internal user_id field is exposed")
		}
	}
}

func TestSortFallsBackToDefault(t *testing.T) {
	defaultSort := Sort("")
	if !defaultSort.Default {
		t.Fatalf("Sort with empty value returned %q, want default", defaultSort.Value)
	}

	unknownSort := Sort("unknown")
	if unknownSort.Value != defaultSort.Value {
		t.Fatalf("Sort with unknown value returned %q, want %q", unknownSort.Value, defaultSort.Value)
	}

	if _, ok := LookupSort("unknown"); ok {
		t.Fatal("LookupSort accepted an unknown value")
	}
}
