// Copyright 2016 The Gogs Authors. All rights reserved.
// Copyright 2019 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"net/http"

	"code.gitea.io/gitea/modules/context"
	"code.gitea.io/gitea/modules/convert"
	"code.gitea.io/gitea/modules/git"
	api "code.gitea.io/gitea/modules/structs"
)

// GetBranch get a branch of a repository
func GetBranch(ctx *context.APIContext) {
	// swagger:operation GET /repos/{owner}/{repo}/branches/{branch} repository repoGetBranch
	// ---
	// summary: Retrieve a specific branch from a repository, including its effective branch protection
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

	branchProtection, err := ctx.Repo.Repository.GetBranchProtection(ctx.Repo.BranchName)
	if err != nil {
		ctx.Error(500, "GetBranchProtection", err)
		return
	}

	ctx.JSON(200, convert.ToBranch(ctx.Repo.Repository, branch, c, branchProtection, ctx.User))
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
		branchProtection, err := ctx.Repo.Repository.GetBranchProtection(branches[i].Name)
		if err != nil {
			ctx.Error(500, "GetBranchProtection", err)
			return
		}
		apiBranches[i] = convert.ToBranch(ctx.Repo.Repository, branches[i], c, branchProtection, ctx.User)
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

	branchProtection, err := ctx.Repo.Repository.GetBranchProtection(ctx.Repo.BranchName)
	if err != nil {
		ctx.Error(500, "GetBranchProtection", err)
		return
	}

	ctx.JSON(200, convert.ToBranch(ctx.Repo.Repository, branch, c, branchProtection, ctx.User))
}





