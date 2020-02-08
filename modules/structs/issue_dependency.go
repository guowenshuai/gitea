package structs

// CreateIssueDependencyOption options for create a dependency
type CreateIssueDependencyOption struct {
	NewDependency int64 `json:"newDependency" binding:"Required"`
}

// RemoveIssueDependencyOption options for remove a dependency
type RemoveIssueDependencyOption struct {
	// removed dependency issue id
	// required: true
	RemoveDependency int64  `json:"removeDependency" binding:"Required"`
	// removed type,  blockedBy or blocking
	// required: true
	DependencyType   string `json:"dependencyType" binding:"Required"`
}
