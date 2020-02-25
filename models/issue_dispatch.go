package models

import (
	"errors"

	"code.gitea.io/gitea/modules/log"
	"code.gitea.io/gitea/modules/util"
)

// IssueDispatch represents an issue dispatch info
type IssueDispatch struct {
	ID            int64          `xorm:"pk autoincr"`
	UserID        int64          `xorm:"NOT NULL"`
	IssueID       int64          `xorm:"UNIQUE NOT NULL"`
	DispatchedID  int64          `xorm:"UNIQUE NOT NULL"` // a issue binding with pr
	RepoID        int64          `xorm:"NOT NULL"`        // dispatch repo
	PullRequestID int64          `xorm:"UNIQUE NOT NULL"` // dispatched pr
	CreatedUnix   util.TimeStamp `xorm:"created"`
	UpdatedUnix   util.TimeStamp `xorm:"updated"`
}

type Dispatch struct {
	*PullRequest
	*Repository
	*Issue
}

// CreateDispatch creates a dispatch record
func CreateDispatch(user *User, issue, dispatch *Issue, repo *Repository, pr *PullRequest) error {
	if issue == nil || dispatch == nil {
		return errors.New("issue nil")
	}
	sess := x.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	exist, err := issueDisPatchExists(sess, issue.ID, dispatch.ID, pr.ID)
	if err != nil {
		return err
	}
	if exist {
		return ErrDispatchExists{
			IssueID:       issue.ID,
			DispatchID:    dispatch.ID,
			PullRequestID: pr.ID,
		}
	}

	if _, err := sess.Insert(&IssueDispatch{
		UserID:        user.ID,
		IssueID:       issue.ID,
		DispatchedID:  dispatch.ID,
		RepoID:        repo.ID,
		PullRequestID: pr.ID,
	}); err != nil {
		return err
	}
	return sess.Commit()
}

// Check if the dispatch already exists
func issueDisPatchExists(e Engine, issueID, dispatchID, pullRequestID int64) (bool, error) {
	return e.Where("(issue_id = ?)", issueID).Exist(&IssueDispatch{})
}

// RemoveIssueDispatch removes a group reaction from an issue
func RemoveIssueDispatch(user *User, issue, dispatch *Issue, pr *PullRequest) (err error) {
	sess := x.NewSession()
	defer sess.Close()
	if err = sess.Begin(); err != nil {
		return err
	}

	issueDispatchDelete := IssueDispatch{
		UserID:        user.ID,
		IssueID:       issue.ID,
		DispatchedID:  dispatch.ID,
		PullRequestID: pr.ID,
	}

	affected, err := sess.Delete(&issueDispatchDelete)
	if err != nil {
		return err
	}

	// If we deleted nothing, the dependency did not exist
	if affected <= 0 {
		return ErrDispatchNotExists{
			IssueID:       issue.ID,
			DispatchID:    dispatch.ID,
			PullRequestID: pr.ID,
		}
	}

	return sess.Commit()
}

// GetDispatch return dispatched issues
func (issue *Issue) GetDispatch() (*Dispatch, error) {
	dispatch, err := issue.getDispatch1(x)
	if err != nil {
		log.Error("GetDispatch.getDispatch error: %+v", err)
		return nil, err
	}
	// dispatch := &IssueDispatch{}
	//
	// if len(dispatches) == 1 {
	// 	dispatch = dispatches[0]
	// } else {
	// 	return nil, errors.New("error get dispatch, to match record")
	// }

	// get dispatch pr
	pr, err := getPullRequestByID(x, dispatch.PullRequestID)
	if err != nil {
		log.Error("GetDispatch.getPullRequestByID error: %+v", err)
		return nil, err
	}
	// get dispatched repo
	repo, err := getRepositoryByID(x, dispatch.RepoID)
	if err != nil {
		log.Error("GetDispatch.getRepositoryByID error: %+v", err)
		return nil, err
	}
	// get dispatched issue
	diss, err := getIssueByID(x, dispatch.DispatchedID)
	if err != nil {
		log.Error("GetDispatch.getIssueByID error: %+v", err)
		return nil, err
	}
	return &Dispatch{
		Repository:  repo,
		PullRequest: pr,
		Issue:       diss,
	}, nil

	// return issue.getDispatch(x)
}

func (issue *Issue) getDispatch(e Engine) (dispatch []*IssueDispatch, err error) {
	return dispatch, e.Where("issue_id = ?", issue.ID).Find(&dispatch)
}
func (issue *Issue) getDispatch1(e Engine) (dispatch *IssueDispatch, err error) {
	d := &IssueDispatch{
		IssueID: issue.ID,
	}
	has, err := e.Get(d)
	if err != nil {
		return nil, err
	} else if !has {
		return nil, ErrDispatchNotExists{
			IssueID: issue.ID,
		}
	}
	return d, nil

}
