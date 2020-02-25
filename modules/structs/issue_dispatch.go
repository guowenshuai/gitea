package structs

// CreateIssueDispatchOption options for create a dependency
type CreateIssueDispatchOption struct {
	// repository id which will dispatched
	// required: true
	Repository int64 `json:"repository" binding:"Required"` //  repository id which will dispatched
	// base branch, default is master
	// required: false
	Base string `json:"base" binding:"MaxSize(255)"`
}

type Dispatch struct {
	Repo  *Repository  `json:"repo"`
	Pr    *PullRequest `json:"pr"`
	Issue *Issue       `json:"issue"`
}

// // RemoveIssueDispatchOption options for remove a dependency
// type RemoveIssueDispatchOption struct {
// 	// removed dependency issue id
// 	// required: true
// 	RemoveDependency int64 `json:"removeDependency" binding:"Required"`
// 	// removed type,  blockedBy or blocking
// 	// required: true
// 	DependencyType string `json:"dependencyType" binding:"Required"`
// }
