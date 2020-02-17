package repo

import (
	"net/http"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/context"
	api "code.gitea.io/gitea/modules/structs"
)

// AddDependency add dependency for issue
func AddDependency(ctx *context.APIContext, form api.CreateIssueDependencyOption) {
	// swagger:operation Post /repos/{owner}/{repo}/issues/{index}/dependency issue issueAddDependency
	// ---
	// summary: Add a Dependency to a issue
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: index
	//   in: path
	//   description: index of the issue
	//   type: integer
	//   format: int64
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/CreateIssueDependencyOption"
	// responses:
	//   "201":
	//     "$ref": "#/responses/Issue"

	// Check if the Repo is allowed to have dependencies
	if !ctx.Repo.CanCreateIssueDependencies(ctx.User) {
		ctx.Error(http.StatusForbidden, "AddDependency", "Can Create Issue Dependencies")
		return
	}

	depID := form.NewDependency

	issueIndex := ctx.ParamsInt64("index")
	issue, err := models.GetIssueByIndex(ctx.Repo.Repository.ID, issueIndex)
	if err != nil {
		ctx.Error(http.StatusNotFound, "GetIssueByIndex", err)
		return
	}

	// Dependency
	dep, err := models.GetIssueByID(depID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "GetIssueByID", err)
		return
	}

	// Check if both issues are in the same repo
	if issue.RepoID != dep.RepoID {
		ctx.Error(http.StatusInternalServerError, "repo.issues.dependency.add_error_dep_not_same_repo", "add_error_dep_not_same_repo")
		return
	}

	// Check if issue and dependency is the same
	if dep.Index == issueIndex {
		ctx.Error(http.StatusInternalServerError, "repo.issues.dependency.add_error_same_issue", "add_error_same_issue")
		return
	}

	err = models.CreateIssueDependency(ctx.User, issue, dep)
	if err != nil {
		if models.IsErrDependencyExists(err) {
			ctx.Error(http.StatusForbidden, "repo.issues.dependency.add_error_dep_exists", err)
			return
		} else if models.IsErrCircularDependency(err) {
			ctx.Error(http.StatusForbidden, "repo.issues.dependency.add_error_cannot_create_circular", err)
			return
		} else {
			ctx.Error(http.StatusInternalServerError, "CreateOrUpdateIssueDependency", err)
			return
		}
	}
	GetIssue(ctx)
}

// RemoveDependency removes the dependency
func RemoveDependency(ctx *context.APIContext, form api.RemoveIssueDependencyOption) {
	// swagger:operation Delete /repos/{owner}/{repo}/issues/{index}/dependency issue issueRemoveDependency
	// ---
	// summary: remove a Dependency from a issue
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: owner
	//   in: path
	//   description: owner of the repo
	//   type: string
	//   required: true
	// - name: repo
	//   in: path
	//   description: name of the repo
	//   type: string
	//   required: true
	// - name: index
	//   in: path
	//   description: index of the issue
	//   type: integer
	//   format: int64
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/RemoveIssueDependencyOption"
	// responses:
	//   "201":
	//     "$ref": "#/responses/Issue"

	// Check if the Repo is allowed to have dependencies
	if !ctx.Repo.CanCreateIssueDependencies(ctx.User) {
		ctx.Error(http.StatusForbidden, "AddDependency", "Can Create Issue Dependencies")
		return
	}

	depID := form.RemoveDependency

	issueIndex := ctx.ParamsInt64("index")
	issue, err := models.GetIssueByIndex(ctx.Repo.Repository.ID, issueIndex)
	if err != nil {
		ctx.Error(http.StatusNotFound, "GetIssueByIndex", err)
		return
	}

	// Dependency Type
	depTypeStr := form.DependencyType

	var depType models.DependencyType

	switch depTypeStr {
	case "blockedBy":
		depType = models.DependencyTypeBlockedBy
	case "blocking":
		depType = models.DependencyTypeBlocking
	default:
		ctx.Error(http.StatusNotFound, "GetDependecyType", "invalid depType")
		return
	}

	// Dependency
	dep, err := models.GetIssueByID(depID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "GetIssueByID", err)
		return
	}

	if err = models.RemoveIssueDependency(ctx.User, issue, dep, depType); err != nil {
		if models.IsErrDependencyNotExists(err) {
			ctx.Error(http.StatusBadRequest, "repo.issues.dependency.remove_error_dep_not_exist", err)
			return
		}
		ctx.Error(http.StatusInternalServerError, "RemoveIssueDependency", err)
		return
	}
	GetIssue(ctx)
}
