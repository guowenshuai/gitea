// Copyright 2016 The Gogs Authors. All rights reserved.
// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"net/http"

	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/git"
	"code.gitea.io/gitea/modules/log"
	api "code.gitea.io/gitea/modules/structs"
	"code.gitea.io/gitea/routers/api/v1/convert"
)

// GetBranch get a branch of a repository
func GetBranch(ctx *context.APIContext) {
	// swagger:operation GET /repos/{owner}/{repo}/branches/{branch} repository repoGetBranch
	// ---
	// summary: Retrieve a specific branch from a repository
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
	// - name: branch
	//   in: path
	//   description: branch to get
	//   type: string
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/Branch"
	if ctx.Repo.TreePath != "" {
		// if TreePath != "", then URL contained extra slashes
		// (i.e. "master/subbranch" instead of "master"), so branch does
		// not exist
		ctx.NotFound()
		return
	}
	branch, err := ctx.Repo.Repository.GetBranch(ctx.Repo.BranchName)
	if err != nil {
		if git.IsErrBranchNotExist(err) {
			ctx.NotFound(err)
		} else {
			ctx.Error(500, "GetBranch", err)
		}
		return
	}

	c, err := branch.GetCommit()
	if err != nil {
		ctx.Error(500, "GetCommit", err)
		return
	}

	ctx.JSON(200, convert.ToBranch(ctx.Repo.Repository, branch, c))
}

// ListBranches list all the branches of a repository
func ListBranches(ctx *context.APIContext) {
	// swagger:operation GET /repos/{owner}/{repo}/branches repository repoListBranches
	// ---
	// summary: List a repository's branches
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
	// responses:
	//   "200":
	//     "$ref": "#/responses/BranchList"
	branches, err := ctx.Repo.Repository.GetBranches()
	if err != nil {
		ctx.Error(500, "GetBranches", err)
		return
	}

	apiBranches := make([]*api.Branch, len(branches))
	for i := range branches {
		c, err := branches[i].GetCommit()
		if err != nil {
			ctx.Error(500, "GetCommit", err)
			return
		}
		apiBranches[i] = convert.ToBranch(ctx.Repo.Repository, branches[i], c)
	}

	ctx.JSON(200, &apiBranches)
}


func CreateBranch(ctx *context.APIContext, form api.CreateBranchOption) {
	// swagger:operation POST /repos/{owner}/{repo}/branch repository repoCreateBranch
	// ---
	// summary: Create NewBranch Base OldBranch
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
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/CreateBranchOption"
	// responses:
	//   "200":
	//     "$ref": "#/responses/Branch"

	if !ctx.Repo.CanCreateBranch() {
		ctx.NotFound("CreateBranch", nil)
		return
	}

	err := ctx.Repo.Repository.CreateNewBranch(ctx.User, form.OldBranch, form.NewBranch)
	if err != nil {
		ctx.Error(http.StatusInternalServerError, "CreateNewBranch", err)
		return
	}

	// get branch
	branch, err := ctx.Repo.Repository.GetBranch(form.NewBranch)
	if err != nil {
		if git.IsErrBranchNotExist(err) {
			ctx.NotFound(err)
		} else {
			ctx.Error(500, "GetBranch", err)
		}
		return
	}

	c, err := branch.GetCommit()
	if err != nil {
		ctx.Error(500, "GetCommit", err)
		return
	}

	ctx.JSON(200, convert.ToBranch(ctx.Repo.Repository, branch, c))
}

func DeleteBranch(ctx *context.APIContext) {
	// swagger:operation DELETE /repos/{owner}/{repo}/branch/{name} repository repoDeleteBranch
	// ---
	// summary: Delete a branch
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
	// - name: name
	//   in: path
	//   description: name of branch
	//   required: true
	// responses:
	//   "200":
	//     "$ref": "#/responses/Branch"

	branchName := ctx.ParamsEscape(":name")
	isProtected, err := ctx.Repo.Repository.IsProtectedBranch(branchName, ctx.User)
	if err != nil {
		log.Error("DeleteBranch: %v", err)
		ctx.Error(http.StatusBadRequest, "repo.branch.deletion_failed", err)
		return
	}

	if isProtected {
		ctx.Error(http.StatusBadRequest, "repo.branch.protected_deletion_failed", err)
		return
	}

	if ctx.Repo.GitRepo == nil {
		ctx.Error(500, "nil gitrepo", "nil git repo")
		return
	}

	// get branch
	branch, err := ctx.Repo.Repository.GetBranch(branchName)
	if err != nil {
		if git.IsErrBranchNotExist(err) {
			ctx.NotFound(err)
		} else {
			ctx.Error(500, "GetBranch", err)
		}
		return
	}

	c, err := branch.GetCommit()
	if err != nil {
		ctx.Error(500, "GetCommit", err)
		return
	}

	if err := ctx.Repo.GitRepo.DeleteBranch(branchName, git.DeleteBranchOptions{
		Force: true,
	}); err != nil {
		log.Error("DeleteBranch: %v", err)
		return
	}

	ctx.JSON(200, convert.ToBranch(ctx.Repo.Repository, branch, c))


}
