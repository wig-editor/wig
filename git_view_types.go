package wig

// GitViewItem represents a single line in the git status view panel.
type GitViewItem struct {
	Type     string // "header", "separator", "empty", "blank", "file", "branch", "stash"
	Label    string
	Status   string // "staged", "unstaged", "untracked", "last_commit", "branch", "active_branch", "stash"
	FilePath string
	Code     string
	StashRef string
}

// CommitItem represents a single commit entry in the git log.
type CommitItem struct {
	Hash    string
	Author  string
	Date    string
	Subject string
}
