package schema

import "strings"

// registry is a map from section name → SectionSchema, built once at startup.
var registry map[string]SectionSchema

func init() {
	registry = make(map[string]SectionSchema, len(All))
	for _, s := range All {
		registry[s.Name] = s
	}
}

// Lookup returns the SectionSchema for the given section name, plus a boolean
// indicating whether a schema was found.
//
// Section names are case-insensitive in git config (e.g. "User" == "user"),
// so we normalise to lowercase before looking up.
//
//	s, ok := schema.Lookup("core")   // ok == true
//	s, ok := schema.Lookup("foobar") // ok == false → use Custom block
func Lookup(name string) (SectionSchema, bool) {
	s, ok := registry[strings.ToLower(name)]
	return s, ok
}

// CommonFields returns the fields from s that are marked Common == true,
// in their original order. Used to populate the visible (non-collapsed) field
// list when a section block is first dropped onto the canvas.
func CommonFields(s SectionSchema) []Field {
	var out []Field
	for _, f := range s.Fields {
		if f.Common {
			out = append(out, f)
		}
	}
	return out
}

// UncommonFields returns fields where Common == false.
// These appear in the [+ Add field] picker under a section divider.
func UncommonFields(s SectionSchema) []Field {
	var out []Field
	for _, f := range s.Fields {
		if !f.Common {
			out = append(out, f)
		}
	}
	return out
}

// FieldByKey returns the Field with the given key from s, or an empty Field
// and false if the key is not in the schema.
//
// Used when loading an existing config: for each key=value pair we look up
// the field schema to know which input widget to render.
func FieldByKey(s SectionSchema, key string) (Field, bool) {
	lower := strings.ToLower(key)
	for _, f := range s.Fields {
		if strings.ToLower(f.Key) == lower {
			return f, true
		}
	}
	return Field{}, false
}

// ConditionTypes returns the list of valid condition type strings for an
// [includeIf] header, in the order they should appear in the dropdown.
//
// These are the part before the colon in [includeIf "gitdir:~/work/"].
func ConditionTypes() []string {
	return []string{
		"gitdir",
		"gitdir/i",
		"onbranch",
		"hasconfig",
		"worktree",
	}
}

// ConditionDocs returns a one-line description for a condition type.
// Used as tooltip text in the conditional block UI.
func ConditionDocs(condType string) string {
	switch condType {
	case "gitdir":
		return "Matches when the repo's .git directory path matches a glob pattern, e.g. ~/work/**"
	case "gitdir/i":
		return "Like gitdir but case-insensitive. Useful on macOS and Windows."
	case "onbranch":
		return "Matches when the currently checked-out branch name matches a glob pattern."
	case "hasconfig":
		return "Matches when another config variable already has a specific value (git 2.36+)."
	case "worktree":
		return "Matches based on the worktree path. Similar to gitdir but for git worktree users."
	default:
		return ""
	}
}
