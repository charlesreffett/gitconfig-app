package schema

import "testing"

func TestLookup_KnownSection(t *testing.T) {
	s, ok := Lookup("user")
	if !ok {
		t.Fatal("Lookup(\"user\") returned ok=false")
	}
	if s.Name != "user" {
		t.Errorf("Name = %q; want \"user\"", s.Name)
	}
}

func TestLookup_CaseInsensitive(t *testing.T) {
	_, ok1 := Lookup("Core")
	_, ok2 := Lookup("CORE")
	_, ok3 := Lookup("core")
	if !ok1 || !ok2 || !ok3 {
		t.Error("Lookup should be case-insensitive")
	}
}

func TestLookup_Unknown(t *testing.T) {
	_, ok := Lookup("completelyunknownsection")
	if ok {
		t.Error("Lookup of unknown section should return ok=false")
	}
}

func TestAllSectionsHaveIcons(t *testing.T) {
	for _, s := range All {
		if s.Icon == "" {
			t.Errorf("section %q has no icon", s.Name)
		}
	}
}

func TestAllSectionsRegistered(t *testing.T) {
	for _, s := range All {
		if _, ok := Lookup(s.Name); !ok {
			t.Errorf("section %q is in All but not in registry", s.Name)
		}
	}
}

func TestCommonFields(t *testing.T) {
	common := CommonFields(User)
	if len(common) == 0 {
		t.Fatal("user section should have common fields")
	}
	for _, f := range common {
		if !f.Common {
			t.Errorf("CommonFields returned non-common field %q", f.Key)
		}
	}
}

func TestUncommonFields(t *testing.T) {
	uncommon := UncommonFields(Core)
	for _, f := range uncommon {
		if f.Common {
			t.Errorf("UncommonFields returned common field %q", f.Key)
		}
	}
}

func TestFieldByKey_Found(t *testing.T) {
	f, ok := FieldByKey(User, "email")
	if !ok {
		t.Fatal("FieldByKey(User, \"email\") returned ok=false")
	}
	if f.Key != "email" {
		t.Errorf("Key = %q; want \"email\"", f.Key)
	}
	if f.Type != FieldText {
		t.Errorf("Type = %v; want FieldText", f.Type)
	}
}

func TestFieldByKey_NotFound(t *testing.T) {
	_, ok := FieldByKey(User, "nonexistent")
	if ok {
		t.Error("FieldByKey should return ok=false for unknown key")
	}
}

func TestFieldByKey_CaseInsensitive(t *testing.T) {
	_, ok := FieldByKey(Core, "AutoCRLF")
	if !ok {
		t.Error("FieldByKey should be case-insensitive")
	}
}

func TestDropdownFieldsHaveOptions(t *testing.T) {
	for _, sec := range All {
		for _, f := range sec.Fields {
			if f.Type == FieldDropdown && len(f.Options) == 0 {
				t.Errorf("section %q field %q is FieldDropdown but has no Options", sec.Name, f.Key)
			}
		}
	}
}

func TestConditionTypes(t *testing.T) {
	types := ConditionTypes()
	if len(types) == 0 {
		t.Fatal("ConditionTypes returned empty slice")
	}
	// gitdir must be first — it's the most common
	if types[0] != "gitdir" {
		t.Errorf("first condition type = %q; want \"gitdir\"", types[0])
	}
}

func TestConditionDocs_AllTypesHaveDocs(t *testing.T) {
	for _, ct := range ConditionTypes() {
		if ConditionDocs(ct) == "" {
			t.Errorf("ConditionDocs(%q) returned empty string", ct)
		}
	}
}

func TestSubsectionSectionsHaveLabels(t *testing.T) {
	for _, s := range All {
		if s.HasSubsection {
			if s.SubsectionLabel == "" {
				t.Errorf("section %q HasSubsection=true but SubsectionLabel is empty", s.Name)
			}
			if s.SubsectionPlaceholder == "" {
				t.Errorf("section %q HasSubsection=true but SubsectionPlaceholder is empty", s.Name)
			}
		}
	}
}
