package schema

// All defines the schema for every known git config section type.
// The order here determines the order in the block library panel.
//
// Sources: git-config(1), git documentation at https://git-scm.com/docs/git-config
var All = []SectionSchema{
	User,
	Core,
	URL,
	Filter,
	Alias,
	HTTP,
	Push,
	Pull,
	Fetch,
	Branch,
	Merge,
	Diff,
	Commit,
	Remote,
	Rebase,
	Color,
	Tag,
	Credential,
	Submodule,
	Init,
	GPG,
}

// ── Identity ──────────────────────────────────────────────────────────────────

var User = SectionSchema{
	Name:  "user",
	Label: "User",
	Icon:  "👤",
	Fields: []Field{
		{
			Key:    "name",
			Label:  "Name",
			Type:   FieldText,
			Common: true,
			Docs:   "Your full name, shown in commit author/committer fields.",
		},
		{
			Key:    "email",
			Label:  "Email",
			Type:   FieldText,
			Common: true,
			Docs:   "Your email address, shown in commit author/committer fields.",
		},
		{
			Key:    "signingkey",
			Label:  "Signing key",
			Type:   FieldText,
			Common: false,
			Docs:   "GPG key ID or SSH key path used to sign commits and tags.",
		},
		{
			Key:     "useConfigOnly",
			Label:   "Use config only",
			Type:    FieldToggle,
			Default: "false",
			Common:  false,
			Docs:    "If true, git refuses to auto-detect your name/email from the system. Forces you to always set them explicitly.",
		},
	},
}

// ── Core behaviour ────────────────────────────────────────────────────────────

var Core = SectionSchema{
	Name:  "core",
	Label: "Core",
	Icon:  "🔧",
	Fields: []Field{
		{
			Key:     "autocrlf",
			Label:   "Line ending conversion",
			Type:    FieldDropdown,
			Options: []string{"input", "true", "false"},
			Default: "false",
			Common:  true,
			Docs:    "'input' converts CRLF→LF on commit only (recommended on macOS/Linux). 'true' also converts LF→CRLF on checkout (Windows). 'false' disables conversion.",
		},
		{
			Key:    "editor",
			Label:  "Editor",
			Type:   FieldText,
			Common: true,
			Docs:   "The editor launched for commit messages, interactive rebases, etc. e.g. 'vim', 'nano', 'code --wait'.",
		},
		{
			Key:    "excludesfile",
			Label:  "Global .gitignore file",
			Type:   FieldText,
			Common: true,
			Docs:   "Path to a global gitignore file applied to all repos, e.g. ~/.gitignore_global.",
		},
		{
			Key:    "hooksPath",
			Label:  "Hooks directory",
			Type:   FieldText,
			Common: false,
			Docs:   "Override the default hooks directory (.git/hooks). Useful for sharing hooks across repos.",
		},
		{
			Key:     "fileMode",
			Label:   "Track file permissions",
			Type:    FieldToggle,
			Default: "true",
			Common:  false,
			Docs:    "If false, git ignores executable bit changes on files. Set to false on Windows/WSL where file modes are unreliable.",
		},
		{
			Key:     "ignorecase",
			Label:   "Ignore filename case",
			Type:    FieldToggle,
			Default: "false",
			Common:  false,
			Docs:    "If true, git treats filenames as case-insensitive. Set automatically on macOS (HFS+) and Windows.",
		},
		{
			Key:     "symlinks",
			Label:   "Follow symlinks",
			Type:    FieldToggle,
			Default: "true",
			Common:  false,
			Docs:    "If false, git checks out symlinks as small plain files containing the link text.",
		},
		{
			Key:    "pager",
			Label:  "Pager",
			Type:   FieldText,
			Common: false,
			Docs:   "The pager used for long output (log, diff, etc.). Defaults to 'less'. Set to '' to disable paging.",
		},
		{
			Key:     "whitespace",
			Label:   "Whitespace checks",
			Type:    FieldText,
			Common:  false,
			Docs:    "Comma-separated list of whitespace rules to enforce (e.g. 'trailing-space,space-before-tab').",
		},
		{
			Key:     "compression",
			Label:   "Compression level",
			Type:    FieldDropdown,
			Options: []string{"-1", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9"},
			Default: "-1",
			Common:  false,
			Docs:    "Zlib compression level for objects. -1 uses the zlib default. 0 disables compression.",
		},
		{
			Key:     "logAllRefUpdates",
			Label:   "Log all ref updates",
			Type:    FieldDropdown,
			Options: []string{"true", "false", "always"},
			Default: "true",
			Common:  false,
			Docs:    "Controls whether a reflog entry is created when a ref is updated.",
		},
		{
			Key:    "attributesfile",
			Label:  "Global .gitattributes file",
			Type:   FieldText,
			Common: false,
			Docs:   "Path to a global gitattributes file applied to all repos.",
		},
		{
			Key:     "quotepath",
			Label:   "Quote special paths",
			Type:    FieldToggle,
			Default: "true",
			Common:  false,
			Docs:    "If false, file paths with non-ASCII characters are printed as-is rather than quoted.",
		},
		{
			Key:    "bigFileThreshold",
			Label:  "Large file threshold",
			Type:   FieldText,
			Common: false,
			Docs:   "Files larger than this value are not delta-compressed. e.g. '512m'. Default is 512 MiB.",
		},
	},
}

// ── URL rewriting ─────────────────────────────────────────────────────────────

var URL = SectionSchema{
	Name:                  "url",
	Label:                 "URL rewrite",
	Icon:                  "🔗",
	HasSubsection:         true,
	SubsectionLabel:       "Target URL",
	SubsectionPlaceholder: "e.g. git@github.com-work",
	Fields: []Field{
		{
			Key:    "insteadOf",
			Label:  "Rewrite from",
			Type:   FieldText,
			Common: true,
			Docs:   "Any URL that begins with this string is rewritten to use the subsection URL instead. e.g. 'git@github.com' → 'git@github.com-work'.",
		},
		{
			Key:    "pushInsteadOf",
			Label:  "Push: rewrite from",
			Type:   FieldText,
			Common: false,
			Docs:   "Like insteadOf but only applies to push operations. Useful for fetching via HTTPS but pushing via SSH.",
		},
	},
}

// ── Filters (e.g. Git LFS) ────────────────────────────────────────────────────

var Filter = SectionSchema{
	Name:                  "filter",
	Label:                 "Filter",
	Icon:                  "📦",
	HasSubsection:         true,
	SubsectionLabel:       "Filter name",
	SubsectionPlaceholder: "e.g. lfs",
	Fields: []Field{
		{
			Key:    "process",
			Label:  "Long-running process",
			Type:   FieldText,
			Common: true,
			Docs:   "A long-running filter process command. Used by Git LFS: 'git-lfs filter-process'.",
		},
		{
			Key:     "required",
			Label:   "Required",
			Type:    FieldToggle,
			Default: "false",
			Common:  true,
			Docs:    "If true, git fails with an error if the filter command is not found. Prevents silently committing unfiltered data.",
		},
		{
			Key:    "clean",
			Label:  "Clean command",
			Type:   FieldText,
			Common: false,
			Docs:   "Command run when staging a file (working tree → index). Used by LFS: 'git-lfs clean -- %f'.",
		},
		{
			Key:    "smudge",
			Label:  "Smudge command",
			Type:   FieldText,
			Common: false,
			Docs:   "Command run when checking out a file (index → working tree). Used by LFS: 'git-lfs smudge -- %f'.",
		},
	},
}

// ── Aliases ───────────────────────────────────────────────────────────────────

var Alias = SectionSchema{
	Name:     "alias",
	Label:    "Aliases",
	Icon:     "🏷️",
	FreeForm: true,
	Fields:   nil, // all entries are user-defined; no fixed schema
}

// ── Network ───────────────────────────────────────────────────────────────────

var HTTP = SectionSchema{
	Name:  "http",
	Label: "HTTP",
	Icon:  "🌐",
	Fields: []Field{
		{
			Key:    "proxy",
			Label:  "Proxy",
			Type:   FieldText,
			Common: true,
			Docs:   "HTTP proxy URL for all HTTP/HTTPS operations, e.g. 'http://proxy.example.com:8080'.",
		},
		{
			Key:     "sslVerify",
			Label:   "Verify SSL certificates",
			Type:    FieldToggle,
			Default: "true",
			Common:  true,
			Docs:    "If false, git skips SSL certificate verification. Only disable this for testing — it opens you to MITM attacks.",
		},
		{
			Key:    "cookieFile",
			Label:  "Cookie file",
			Type:   FieldText,
			Common: false,
			Docs:   "Path to a Netscape-format cookie file sent with HTTP requests.",
		},
		{
			Key:     "saveCookies",
			Label:   "Save cookies",
			Type:    FieldToggle,
			Default: "false",
			Common:  false,
			Docs:    "If true, save cookies received from the server back to cookieFile.",
		},
		{
			Key:     "version",
			Label:   "HTTP version",
			Type:    FieldDropdown,
			Options: []string{"HTTP/1.1", "HTTP/2"},
			Common:  false,
			Docs:    "HTTP protocol version to use. HTTP/2 requires a libcurl built with HTTP/2 support.",
		},
		{
			Key:    "sslCert",
			Label:  "SSL client certificate",
			Type:   FieldText,
			Common: false,
			Docs:   "Path to the SSL client certificate file.",
		},
		{
			Key:    "sslKey",
			Label:  "SSL client key",
			Type:   FieldText,
			Common: false,
			Docs:   "Path to the SSL client key file.",
		},
		{
			Key:    "sslCAInfo",
			Label:  "SSL CA bundle",
			Type:   FieldText,
			Common: false,
			Docs:   "Path to a file containing SSL CA certificates used to verify the server.",
		},
		{
			Key:     "followRedirects",
			Label:   "Follow redirects",
			Type:    FieldDropdown,
			Options: []string{"true", "false", "initial"},
			Default: "initial",
			Common:  false,
			Docs:    "'initial' follows redirects only on the first request. 'true' follows all. 'false' never follows.",
		},
		{
			Key:    "extraHeader",
			Label:  "Extra header",
			Type:   FieldText,
			Common: false,
			Docs:   "An additional HTTP header to send with every request, e.g. 'Authorization: Bearer token'.",
		},
	},
}

// ── Push / Pull / Fetch ───────────────────────────────────────────────────────

var Push = SectionSchema{
	Name:  "push",
	Label: "Push",
	Icon:  "⬆️",
	Fields: []Field{
		{
			Key:     "default",
			Label:   "Default push behaviour",
			Type:    FieldDropdown,
			Options: []string{"simple", "current", "upstream", "matching", "nothing"},
			Default: "simple",
			Common:  true,
			Docs:    "'simple' (recommended) pushes to the upstream branch with the same name. 'current' pushes to same-named remote branch. 'upstream' pushes to whatever the local branch tracks. 'matching' pushes all matching branches.",
		},
		{
			Key:     "followTags",
			Label:   "Push tags with commits",
			Type:    FieldToggle,
			Default: "false",
			Common:  true,
			Docs:    "If true, automatically push annotated tags that are reachable from pushed commits.",
		},
		{
			Key:     "autoSetupRemote",
			Label:   "Auto-setup remote tracking",
			Type:    FieldToggle,
			Default: "false",
			Common:  false,
			Docs:    "If true, 'git push' automatically sets up remote tracking for new branches (--set-upstream).",
		},
		{
			Key:     "recurseSubmodules",
			Label:   "Push submodules",
			Type:    FieldDropdown,
			Options: []string{"no", "check", "on-demand"},
			Default: "no",
			Common:  false,
			Docs:    "'check' aborts if submodules have unpushed commits. 'on-demand' pushes submodules automatically.",
		},
	},
}

var Pull = SectionSchema{
	Name:  "pull",
	Label: "Pull",
	Icon:  "⬇️",
	Fields: []Field{
		{
			Key:     "rebase",
			Label:   "Pull strategy",
			Type:    FieldDropdown,
			Options: []string{"false", "true", "merges", "interactive"},
			Default: "false",
			Common:  true,
			Docs:    "'false' merges the fetched branch. 'true' rebases on top of it. 'merges' preserves merge commits during rebase. 'interactive' opens an interactive rebase.",
		},
		{
			Key:     "ff",
			Label:   "Fast-forward policy",
			Type:    FieldDropdown,
			Options: []string{"true", "false", "only"},
			Default: "true",
			Common:  true,
			Docs:    "'true' fast-forwards when possible, otherwise merges. 'false' always creates a merge commit. 'only' fails if fast-forward is not possible.",
		},
	},
}

var Fetch = SectionSchema{
	Name:  "fetch",
	Label: "Fetch",
	Icon:  "🔄",
	Fields: []Field{
		{
			Key:     "prune",
			Label:   "Auto-prune deleted remote branches",
			Type:    FieldToggle,
			Default: "false",
			Common:  true,
			Docs:    "If true, 'git fetch' automatically removes local remote-tracking branches that no longer exist on the remote.",
		},
		{
			Key:     "pruneTags",
			Label:   "Auto-prune deleted remote tags",
			Type:    FieldToggle,
			Default: "false",
			Common:  false,
			Docs:    "If true, also prune local tags that no longer exist on the remote. Requires fetch.prune to be true.",
		},
		{
			Key:     "all",
			Label:   "Fetch all remotes",
			Type:    FieldToggle,
			Default: "false",
			Common:  false,
			Docs:    "If true, 'git fetch' fetches from all remotes, not just origin.",
		},
		{
			Key:     "recurseSubmodules",
			Label:   "Fetch submodules",
			Type:    FieldDropdown,
			Options: []string{"on-demand", "yes", "no"},
			Default: "on-demand",
			Common:  false,
			Docs:    "Controls whether submodule commits reachable from the superproject are also fetched.",
		},
		{
			Key:    "parallel",
			Label:  "Parallel fetch jobs",
			Type:   FieldText,
			Common: false,
			Docs:   "Number of parallel child processes to fetch multiple remotes or submodules simultaneously. 0 = unlimited.",
		},
	},
}

// ── Branch / Remote ───────────────────────────────────────────────────────────

var Branch = SectionSchema{
	Name:                  "branch",
	Label:                 "Branch",
	Icon:                  "🌿",
	HasSubsection:         true,
	SubsectionLabel:       "Branch name",
	SubsectionPlaceholder: "e.g. main",
	Fields: []Field{
		{
			Key:    "remote",
			Label:  "Tracking remote",
			Type:   FieldText,
			Common: true,
			Docs:   "The remote this branch tracks, e.g. 'origin'.",
		},
		{
			Key:    "merge",
			Label:  "Tracking ref",
			Type:   FieldText,
			Common: true,
			Docs:   "The remote ref this branch merges from/pushes to, e.g. 'refs/heads/main'.",
		},
		{
			Key:     "rebase",
			Label:   "Rebase on pull",
			Type:    FieldDropdown,
			Options: []string{"false", "true", "interactive", "merges"},
			Default: "false",
			Common:  false,
			Docs:    "Overrides pull.rebase for this specific branch.",
		},
		{
			Key:    "pushRemote",
			Label:  "Push remote",
			Type:   FieldText,
			Common: false,
			Docs:   "Override the remote to push to for this branch (separate from the tracking remote).",
		},
		{
			Key:     "description",
			Label:   "Description",
			Type:    FieldMultiline,
			Common:  false,
			Docs:    "A description of this branch's purpose. Shown by 'git branch --edit-description'.",
		},
	},
}

var Remote = SectionSchema{
	Name:                  "remote",
	Label:                 "Remote",
	Icon:                  "🏠",
	HasSubsection:         true,
	SubsectionLabel:       "Remote name",
	SubsectionPlaceholder: "e.g. origin",
	Fields: []Field{
		{
			Key:    "url",
			Label:  "URL",
			Type:   FieldText,
			Common: true,
			Docs:   "The remote repository URL, e.g. 'git@github.com:user/repo.git'.",
		},
		{
			Key:    "fetch",
			Label:  "Fetch refspec",
			Type:   FieldText,
			Common: true,
			Docs:   "The refspec for 'git fetch', e.g. '+refs/heads/*:refs/remotes/origin/*'.",
		},
		{
			Key:    "pushurl",
			Label:  "Push URL",
			Type:   FieldText,
			Common: false,
			Docs:   "A separate URL used only for push operations. Falls back to 'url' if not set.",
		},
		{
			Key:    "push",
			Label:  "Push refspec",
			Type:   FieldText,
			Common: false,
			Docs:   "The default refspec for 'git push' to this remote.",
		},
		{
			Key:     "tagOpt",
			Label:   "Tag fetching",
			Type:    FieldDropdown,
			Options: []string{"--tags", "--no-tags"},
			Common:  false,
			Docs:    "'--tags' fetches all tags. '--no-tags' fetches no tags automatically.",
		},
		{
			Key:     "mirror",
			Label:   "Mirror mode",
			Type:    FieldToggle,
			Default: "false",
			Common:  false,
			Docs:    "If true, refs are mirrored exactly from the remote (push --mirror behaviour).",
		},
	},
}

// ── Merge / Diff / Rebase ─────────────────────────────────────────────────────

var Merge = SectionSchema{
	Name:  "merge",
	Label: "Merge",
	Icon:  "🔀",
	Fields: []Field{
		{
			Key:    "tool",
			Label:  "Merge tool",
			Type:   FieldText,
			Common: true,
			Docs:   "The merge tool launched by 'git mergetool', e.g. 'vimdiff', 'vscode', 'opendiff'.",
		},
		{
			Key:     "conflictstyle",
			Label:   "Conflict style",
			Type:    FieldDropdown,
			Options: []string{"merge", "diff3", "zdiff3"},
			Default: "merge",
			Common:  true,
			Docs:    "'merge' shows the two sides. 'diff3' also shows the common ancestor. 'zdiff3' like diff3 but trims common lines.",
		},
		{
			Key:     "ff",
			Label:   "Fast-forward merges",
			Type:    FieldDropdown,
			Options: []string{"true", "false", "only"},
			Default: "true",
			Common:  false,
			Docs:    "'false' always creates a merge commit even when fast-forward is possible.",
		},
		{
			Key:     "stat",
			Label:   "Show diffstat after merge",
			Type:    FieldToggle,
			Default: "true",
			Common:  false,
			Docs:    "If false, suppresses the diffstat summary printed after a merge.",
		},
		{
			Key:     "commit",
			Label:   "Auto-commit after merge",
			Type:    FieldToggle,
			Default: "true",
			Common:  false,
			Docs:    "If false, 'git merge' stops before committing so you can inspect or amend the result.",
		},
		{
			Key:     "edit",
			Label:   "Edit merge commit message",
			Type:    FieldToggle,
			Default: "true",
			Common:  false,
			Docs:    "If false, git uses the auto-generated merge message without opening an editor.",
		},
	},
}

var Diff = SectionSchema{
	Name:  "diff",
	Label: "Diff",
	Icon:  "📊",
	Fields: []Field{
		{
			Key:    "tool",
			Label:  "Diff tool",
			Type:   FieldText,
			Common: true,
			Docs:   "The diff tool launched by 'git difftool', e.g. 'vimdiff', 'vscode', 'kaleidoscope'.",
		},
		{
			Key:     "algorithm",
			Label:   "Diff algorithm",
			Type:    FieldDropdown,
			Options: []string{"default", "minimal", "patience", "histogram"},
			Default: "default",
			Common:  true,
			Docs:    "'patience' and 'histogram' often produce more readable diffs for code. 'histogram' is generally the best choice.",
		},
		{
			Key:     "colorMoved",
			Label:   "Highlight moved lines",
			Type:    FieldDropdown,
			Options: []string{"no", "default", "blocks", "zebra", "dimmed-zebra"},
			Default: "no",
			Common:  false,
			Docs:    "Colours lines that were moved (not just added/removed) differently. 'dimmed-zebra' is the most readable.",
		},
		{
			Key:     "renames",
			Label:   "Detect renames",
			Type:    FieldDropdown,
			Options: []string{"true", "false", "copy"},
			Default: "true",
			Common:  false,
			Docs:    "'true' detects renames. 'copy' also detects copies. 'false' disables rename detection.",
		},
		{
			Key:     "indentHeuristic",
			Label:   "Indent heuristic",
			Type:    FieldToggle,
			Default: "true",
			Common:  false,
			Docs:    "Shifts diff hunk boundaries to produce more readable diffs for indented code.",
		},
		{
			Key:     "mnemonicPrefix",
			Label:   "Mnemonic prefixes",
			Type:    FieldToggle,
			Default: "false",
			Common:  false,
			Docs:    "If true, uses 'i/' (index), 'w/' (working tree), 'c/' (commit) prefixes instead of 'a/' and 'b/'.",
		},
	},
}

var Rebase = SectionSchema{
	Name:  "rebase",
	Label: "Rebase",
	Icon:  "🔁",
	Fields: []Field{
		{
			Key:     "autoSquash",
			Label:   "Auto-squash fixup commits",
			Type:    FieldToggle,
			Default: "false",
			Common:  true,
			Docs:    "If true, commits prefixed with 'fixup!' or 'squash!' are automatically moved and squashed in interactive rebases.",
		},
		{
			Key:     "autoStash",
			Label:   "Auto-stash before rebase",
			Type:    FieldToggle,
			Default: "false",
			Common:  true,
			Docs:    "If true, automatically stashes uncommitted changes before rebasing and pops the stash after.",
		},
		{
			Key:     "updateRefs",
			Label:   "Update stacked branches",
			Type:    FieldToggle,
			Default: "false",
			Common:  false,
			Docs:    "If true, rebase automatically updates any branches that point to rebased commits (stacked branch workflow).",
		},
		{
			Key:     "missingCommitsCheck",
			Label:   "Check for missing commits",
			Type:    FieldDropdown,
			Options: []string{"ignore", "warn", "error"},
			Default: "ignore",
			Common:  false,
			Docs:    "What to do when commits are removed from the rebase todo list. 'error' prevents accidental commit loss.",
		},
		{
			Key:     "backend",
			Label:   "Rebase backend",
			Type:    FieldDropdown,
			Options: []string{"merge", "apply"},
			Default: "merge",
			Common:  false,
			Docs:    "'merge' is the default modern backend. 'apply' is the legacy patch-based backend.",
		},
	},
}

// ── Commit / Tag ──────────────────────────────────────────────────────────────

var Commit = SectionSchema{
	Name:  "commit",
	Label: "Commit",
	Icon:  "📝",
	Fields: []Field{
		{
			Key:     "gpgSign",
			Label:   "Sign all commits",
			Type:    FieldToggle,
			Default: "false",
			Common:  true,
			Docs:    "If true, every commit is GPG/SSH signed automatically, equivalent to always passing -S.",
		},
		{
			Key:     "verbose",
			Label:   "Verbose commit message",
			Type:    FieldToggle,
			Default: "false",
			Common:  true,
			Docs:    "If true, the diff of staged changes is appended to the commit message template in the editor.",
		},
		{
			Key:    "template",
			Label:  "Commit message template",
			Type:   FieldText,
			Common: false,
			Docs:   "Path to a file used as the initial commit message template.",
		},
		{
			Key:     "status",
			Label:   "Include status in message editor",
			Type:    FieldToggle,
			Default: "true",
			Common:  false,
			Docs:    "If false, the 'git status' output is not included as a comment in the commit message editor.",
		},
		{
			Key:     "cleanup",
			Label:   "Message cleanup",
			Type:    FieldDropdown,
			Options: []string{"default", "strip", "whitespace", "verbatim", "scissors"},
			Default: "default",
			Common:  false,
			Docs:    "How git cleans up the commit message. 'scissors' is useful when the diff is included verbosely.",
		},
	},
}

var Tag = SectionSchema{
	Name:  "tag",
	Label: "Tag",
	Icon:  "🏷️",
	Fields: []Field{
		{
			Key:     "gpgSign",
			Label:   "Sign all tags",
			Type:    FieldToggle,
			Default: "false",
			Common:  true,
			Docs:    "If true, 'git tag' always creates signed tags, equivalent to always passing -s.",
		},
		{
			Key:    "sort",
			Label:  "Sort order",
			Type:   FieldText,
			Common: false,
			Docs:   "Default sort key for 'git tag -l', e.g. '-version:refname' for descending version sort.",
		},
		{
			Key:     "forceSignAnnotated",
			Label:   "Force-sign annotated tags",
			Type:    FieldToggle,
			Default: "false",
			Common:  false,
			Docs:    "If true, annotated tags created with 'git tag -a' are also signed.",
		},
	},
}

// ── Credentials / Auth ────────────────────────────────────────────────────────

var Credential = SectionSchema{
	Name:                  "credential",
	Label:                 "Credential",
	Icon:                  "🔑",
	HasSubsection:         true,
	SubsectionLabel:       "URL scope (optional)",
	SubsectionPlaceholder: "e.g. https://github.com  (leave blank for all URLs)",
	Fields: []Field{
		{
			Key:    "helper",
			Label:  "Credential helper",
			Type:   FieldText,
			Common: true,
			Docs:   "The program used to store/retrieve credentials, e.g. 'osxkeychain', 'manager', 'store'.",
		},
		{
			Key:    "username",
			Label:  "Username",
			Type:   FieldText,
			Common: false,
			Docs:   "Default username for this URL scope.",
		},
		{
			Key:     "useHttpPath",
			Label:   "Scope credentials to URL path",
			Type:    FieldToggle,
			Default: "false",
			Common:  false,
			Docs:    "If true, the full URL path is included when looking up credentials, not just the hostname.",
		},
	},
}

// ── Submodules ────────────────────────────────────────────────────────────────

var Submodule = SectionSchema{
	Name:                  "submodule",
	Label:                 "Submodule",
	Icon:                  "📦",
	HasSubsection:         true,
	SubsectionLabel:       "Submodule name",
	SubsectionPlaceholder: "e.g. vendor/mylib",
	Fields: []Field{
		{
			Key:    "path",
			Label:  "Local path",
			Type:   FieldText,
			Common: true,
			Docs:   "The path within the repo where this submodule is checked out.",
		},
		{
			Key:    "url",
			Label:  "URL",
			Type:   FieldText,
			Common: true,
			Docs:   "The URL of the submodule repository.",
		},
		{
			Key:     "update",
			Label:   "Update strategy",
			Type:    FieldDropdown,
			Options: []string{"checkout", "rebase", "merge", "none"},
			Default: "checkout",
			Common:  false,
			Docs:    "How 'git submodule update' updates the submodule working tree.",
		},
		{
			Key:    "branch",
			Label:  "Tracking branch",
			Type:   FieldText,
			Common: false,
			Docs:   "The remote branch this submodule tracks. Used by 'git submodule update --remote'.",
		},
		{
			Key:     "fetchRecurseSubmodules",
			Label:   "Fetch recursively",
			Type:    FieldDropdown,
			Options: []string{"on-demand", "yes", "no"},
			Default: "on-demand",
			Common:  false,
			Docs:    "Whether to recurse into this submodule when fetching the parent repo.",
		},
		{
			Key:     "ignore",
			Label:   "Ignore changes",
			Type:    FieldDropdown,
			Options: []string{"none", "untracked", "dirty", "all"},
			Default: "none",
			Common:  false,
			Docs:    "What submodule changes 'git status' and 'git diff' report. 'all' hides all submodule changes.",
		},
	},
}

// ── Appearance ───────────────────────────────────────────────────────────────

var Color = SectionSchema{
	Name:  "color",
	Label: "Color",
	Icon:  "🎨",
	Fields: []Field{
		{
			Key:     "ui",
			Label:   "Colorize output",
			Type:    FieldDropdown,
			Options: []string{"auto", "true", "always", "false", "never"},
			Default: "auto",
			Common:  true,
			Docs:    "'auto' enables colour when writing to a terminal. 'always' forces colour even when piped. 'false' disables colour entirely.",
		},
		{
			Key:     "branch",
			Label:   "Colorize branch output",
			Type:    FieldDropdown,
			Options: []string{"auto", "true", "always", "false", "never"},
			Default: "auto",
			Common:  false,
			Docs:    "Controls colour in 'git branch' output. Inherits color.ui when not set.",
		},
		{
			Key:     "diff",
			Label:   "Colorize diff output",
			Type:    FieldDropdown,
			Options: []string{"auto", "true", "always", "false", "never"},
			Default: "auto",
			Common:  false,
			Docs:    "Controls colour in 'git diff' output.",
		},
		{
			Key:     "status",
			Label:   "Colorize status output",
			Type:    FieldDropdown,
			Options: []string{"auto", "true", "always", "false", "never"},
			Default: "auto",
			Common:  false,
			Docs:    "Controls colour in 'git status' output.",
		},
		{
			Key:     "interactive",
			Label:   "Colorize interactive prompts",
			Type:    FieldDropdown,
			Options: []string{"auto", "true", "always", "false", "never"},
			Default: "auto",
			Common:  false,
			Docs:    "Controls colour in interactive commands like 'git add -i'.",
		},
		{
			Key:     "pager",
			Label:   "Colorize pager output",
			Type:    FieldToggle,
			Default: "true",
			Common:  false,
			Docs:    "If true, colour codes are passed through to the pager. Set to false if your pager does not handle ANSI colour.",
		},
	},
}

// ── Misc ──────────────────────────────────────────────────────────────────────

var Init = SectionSchema{
	Name:  "init",
	Label: "Init",
	Icon:  "🌱",
	Fields: []Field{
		{
			Key:     "defaultBranch",
			Label:   "Default branch name",
			Type:    FieldText,
			Default: "master",
			Common:  true,
			Docs:    "The branch name used when initialising a new repository with 'git init'. Set to 'main' to use the modern default.",
		},
	},
}

var GPG = SectionSchema{
	Name:  "gpg",
	Label: "GPG / Signing",
	Icon:  "🔐",
	Fields: []Field{
		{
			Key:     "format",
			Label:   "Signing format",
			Type:    FieldDropdown,
			Options: []string{"openpgp", "x509", "ssh"},
			Default: "openpgp",
			Common:  true,
			Docs:    "'openpgp' uses GPG (classic). 'ssh' uses an SSH key for signing (git 2.34+, simplest to set up). 'x509' uses S/MIME.",
		},
		{
			Key:    "program",
			Label:  "GPG program",
			Type:   FieldText,
			Common: false,
			Docs:   "Path to the GPG program, e.g. '/opt/homebrew/bin/gpg'. Defaults to 'gpg' on PATH.",
		},
	},
}
