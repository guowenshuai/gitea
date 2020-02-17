package structs

// CreateIssueGroupReactionOption options for create a group reaction
type CreateIssueGroupReactionOption struct {
	Child int64 `json:"child" binding:"Required"`
}

// RemoveIssueGroupReactionOption options for remove a group reaction
type RemoveIssueGroupReactionOption struct {
	// removed group issue id
	// required: true
	RemoveChild int64  `json:"remove_child" binding:"Required"`
}
