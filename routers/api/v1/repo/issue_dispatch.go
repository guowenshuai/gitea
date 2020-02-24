package repo

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/git"
	api "code.gitea.io/gitea/modules/structs"
	"github.com/lunny/log"
)

// CreateDispatch dispatch a issue to a repository
func CreateDispatch(ctx *context.APIContext, form api.CreateIssueDispatchOption) {
	// swagger:operation Post /repos/{owner}/{repo}/issues/{index}/dispatch issue issueCreateDispatch
	// ---
	// summary: Create a dispatch to a issue
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
	//     "$ref": "#/definitions/CreateIssueDispatchOption"
	// responses:
	//   "201":
	//     "$ref": "#/responses/Issue"

	// 检查 form.repository 仓库是否存在
	dispatchRepoID := form.Repository
	dispatchRepo, err := models.GetRepositoryByID(dispatchRepoID)
	if err != nil {
		if models.IsErrRepoNotExist(err) {
			ctx.NotFound(err)
		} else {
			ctx.Error(http.StatusInternalServerError, "CreateDispatch", err)
		}
		return
	}

	// 检查仓库是否存在 form.base 这个分支
	baseBranch, err := dispatchRepo.GetBranch(form.Base)
	if err != nil {
		if git.IsErrBranchNotExist(err) {
			ctx.NotFound(err)
		} else {
			ctx.Error(500, "GetBranch", err)
		}
		return
	}

	// 根据当前issue.index+issue.title，生成一个名称,基于form.base创建一个分支
	issueIndex := ctx.ParamsInt64("index")
	issue, err := models.GetIssueByIndex(ctx.Repo.Repository.ID, issueIndex)
	if err != nil {
		ctx.Error(http.StatusNotFound, "GetIssueByIndex", err)
		return
	}
	newBranch := generateBranchName(issue)

	ctx2 := *ctx
	ctx2.Repo.Repository = dispatchRepo

	if !ctx2.Repo.CanCreateBranch() {
		ctx.NotFound("CreateBranch", nil)
		return
	}

	err = ctx2.Repo.Repository.CreateNewBranch(ctx.User, baseBranch.Name, newBranch)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "CreateNewBranch", err)
		return
	}

	// 基于form.base和新分支，创建一个pr
	pr, err := createPullRequest(&ctx2, api.CreatePullRequestOption{
		Head:      newBranch,
		Base:      baseBranch.Name,
		Title:     issue.Title,
		Body:      issue.Title,
		Assignee:  "",
		Assignees: nil,
		Milestone: 0,
		Labels:    nil,
		Deadline:  nil,
	})
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "createPullRequest", err)
		return
	}
	// 创建一条dispatch记录

	err = models.CreateDispatch(ctx.User, issue, pr.Issue, pr)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "CreateDispatch", err)
		return
	}

	// 返回issue
	GetIssue(ctx)
}

func generateBranchName(issue *models.Issue) string {
	branch := strconv.FormatInt(issue.Index,10) + "-"
	if len(issue.Title) > 0 {
		reg := regexp.MustCompile(`\w+`)
		s := reg.FindAllString(issue.Title, 2)
		branch = branch + strings.Join(s, "_")
	}
	log.Infof("generateBranchName [%s]\n", branch)
	return branch
}

