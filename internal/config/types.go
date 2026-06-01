package config

// Config is the in-memory representation of a single gitconfig file.
type Config struct {
	FilePath string    // absolute path to the file on disk
	Sections []Section // ordered list of sections as they appear in the file
}

// Section represents one [header] block in a gitconfig file.
// For a plain header like [user], Name="user" and Subsection="".
// For [includeIf "gitdir:~/work/"], Name="includeIf" and Subsection="gitdir:~/work/".
type Section struct {
	Name       string  // section name, always lowercase in canonical git config
	Subsection string  // optional subsection value (the quoted part); empty if none
	Entries    []Entry // key=value pairs (and comment lines) in this section
}

// Entry is a single key=value pair inside a section.
// If Key is empty, the entry is a comment/blank line preserved verbatim in RawLine.
type Entry struct {
	Key     string // e.g. "name", "email", "path", "autocrlf"
	Value   string // e.g. "KnightYoshi", "~/.gitconfig_work"
	RawLine string // original raw line — used for comment/blank preservation
}
