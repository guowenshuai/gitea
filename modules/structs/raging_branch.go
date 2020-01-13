package structs

// CreateBranchOption options for creating a branch
type CreateBranchOption struct {
	OldBranch string `json:"old_branch" binding:"Required;MaxSize(255)"`
	NewBranch string `json:"new_branch" binding:"Required;MaxSize(255)"`
}

