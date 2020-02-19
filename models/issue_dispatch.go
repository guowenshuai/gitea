package models

import "code.gitea.io/gitea/modules/util"

// IssueDispatch represents an issue dispatch info
type IssueDispatch struct {
	ID            int64          `xorm:"pk autoincr"`
	UserID        int64          `xorm:"NOT NULL"`
	IssueID       int64          `xorm:"UNIQUE NOT NULL"`
	DispatchedID  int64          `xorm:"UNIQUE NOT NULL"` // a issue binding with pr
	PullRequestID int64          `xorm:"UNIQUE NOT NULL"` // dispatched pr
	CreatedUnix   util.TimeStamp `xorm:"created"`
	UpdatedUnix   util.TimeStamp `xorm:"updated"`
}

// CreateDispatch creates a dispatch record
func CreateDispatch(user *User, issue, dispatch *Issue, pr *PullRequest) error {
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
func (issue *Issue) GetDispatch() ([]*Issue, error) {
	return issue.getDispatch(x)
}

func (issue *Issue) getDispatch(e Engine) (issues []*Issue, err error) {
	return issues, e.
		Table("issue_dispatch").
		Select("issue.*").
		Join("INNER", "issue", "issue.id = issue_dispatch.dispatched_id").
		Where("issue_id = ?", issue.ID).
		Find(&issues)
}
