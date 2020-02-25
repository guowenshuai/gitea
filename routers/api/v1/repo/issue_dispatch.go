package repo

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"code.gitea.io/gitea/models"
	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/modules/log"
	api "code.gitea.io/gitea/modules/structs"
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

	// load Repo
	// For API calls.
	if ctx2.Repo.GitRepo == nil {
		repoPath := models.RepoPath(ctx2.Repo.Owner.Name, ctx2.Repo.Repository.Name)
		gitRepo, err := git.OpenRepository(repoPath)
		if err != nil {
			ctx.Error(500, "RepoRef Invalid repo "+repoPath, err)
			return
		}
		ctx2.Repo.GitRepo = gitRepo
		// We opened it, we should close it
		defer func() {
			// If it's been set to nil then assume someone else has closed it.
			if ctx2.Repo.GitRepo != nil {
				ctx2.Repo.GitRepo.Close()
			}
		}()
	}

	if !ctx2.Repo.CanCreateBranch() {
		ctx.NotFound("CreateBranch", nil)
		return
	}

	// 新分支不存在，创建新分支
	_, err = ctx2.Repo.Repository.GetBranch(newBranch)
	if err != nil {
		if git.IsErrBranchNotExist(err) {
			log.Trace("create dispatch branch [%s] exists", newBranch)
			err = ctx2.Repo.Repository.CreateNewBranch(ctx.User, baseBranch.Name, newBranch)
			if err != nil {
				ctx.Error(http.StatusInternalServerError, "CreateNewBranch", err)
				return
			}

		} else {
			ctx.Error(500, "GetBranch", err)
			return
		}
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
	err = models.CreateDispatch(ctx.User, issue, pr.Issue, dispatchRepo, pr)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "CreateDispatch", err)
		return
	}

	// 返回issue
	if iss, err := models.GetIssueWithAttrsByID(issue.ID); err != nil {
		ctx.Error(http.StatusInternalServerError, "CreteaDispatch.getIssue", err)
		return
	} else {
		ctx.JSON(200, iss.APIFormat())
		return
	}
}

func generateBranchName(issue *models.Issue) string {
	branch := strconv.FormatInt(issue.Index, 10) + "-"
	if len(issue.Title) > 0 {
		reg := regexp.MustCompile(`\w+`)
		s := reg.FindAllString(issue.Title, 2)
		branch = branch + strings.Join(s, "_")
	}
	log.Info("generateBranchName [%s]\n", branch)
	return branch
}
