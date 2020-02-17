package repo

import (
	"net/http"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/log"
	api "code.gitea.io/gitea/modules/structs"

	"code.gitea.io/gitea/modules/context"
)

// AddGroupReaction add group reaction for issue
func AddGroupReaction(ctx *context.APIContext, form api.CreateIssueGroupReactionOption) {
	// swagger:operation Post /repos/{owner}/{repo}/issues/{index}/groupReaction issue issueAddGroupReaction
	// ---
	// summary: Add a group reaction to a issue
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
	//     "$ref": "#/definitions/CreateIssueGroupReactionOption"
	// responses:
	//   "201":
	//     "$ref": "#/responses/Issue"

	childID := form.Child

	issueIndex := ctx.ParamsInt64("index")
	issue, err := models.GetIssueByIndex(ctx.Repo.Repository.ID, issueIndex)
	if err != nil {
		ctx.Error(http.StatusNotFound, "GetIssueByIndex", err)
		return
	}
	log.Info("AddGroupReaction: issueIndex %-v, childIndex %-v", issueIndex, childID)

	// child issue
	child, err :=  models.GetIssueByID(childID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "GetIssueByID", err)
		return
	}

	// check same repo
	if issue.RepoID != child.RepoID {
		ctx.Error(http.StatusBadRequest, "repo.issue.groupReaction.add_error_not_same_repo", "add_error_reaction_not_same_repo")
		return
	}

	// check same issue
	if issue.Index == child.Index {
		ctx.Error(http.StatusBadRequest, "repo.issues.dependency.add_error_same_issue", "add_error_same_issue")
		return
	}

	err = models.CreateIssueGroupReaction(ctx.User, issue, child)
	if err != nil {
		if models.IsErrGroupReactionExists(err) {
			ctx.Error(http.StatusForbidden, "repo.issues.group_reaction.add_error_exists", err)
			return
		} else if models.IsErrCircularGroupReaction(err) {
			ctx.Error(http.StatusForbidden, "repo.issues.group_reaction.add_error_circular", err)
			return
		} else {
			ctx.Error(http.StatusInternalServerError, "repo.issues.group_reaction.add_error", err)
			return
		}
	}

	GetIssue(ctx)
}

// RemoveGroupReaction remove the group reaction
func RemoveGroupReaction(ctx *context.APIContext, form api.RemoveIssueGroupReactionOption) {
	// swagger:operation Delete /repos/{owner}/{repo}/issues/{index}/groupReaction issue issueRemoveGroupReaction
	// ---
	// summary: remove a group reaction from a issue
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
	//     "$ref": "#/definitions/RemoveIssueGroupReactionOption"
	// responses:
	//   "201":
	//     "$ref": "#/responses/Issue"

	childID := form.RemoveChild

	issueIndex := ctx.ParamsInt64("index")
	issue, err := models.GetIssueByIndex(ctx.Repo.Repository.ID, issueIndex)
	if err != nil {
		ctx.Error(http.StatusNotFound, "GetIssueByIndex", err)
		return
	}

	// child issue
	child, err :=  models.GetIssueByID(childID)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "GetIssueByID", err)
		return
	}

	if err = models.RemoveIssueGroupReaction(ctx.User, issue, child); err != nil {
		if models.IsErrGroupReactionNotExists(err) {
			ctx.Error(http.StatusBadRequest, "repo.issues.groupReaction.remove_error_gr_not_exist", err)
			return
		}
		ctx.Error(http.StatusInternalServerError, "RemoveIssueGroupReaction", err)
		return
	}
	GetIssue(ctx)
}
