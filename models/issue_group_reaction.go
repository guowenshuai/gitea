// issue's group reaction, parent or child
package models

import (
	"code.gitea.io/gitea/modules/util"
)

// IssueGroupReaction represents an issue group reaction
type IssueGroupReaction struct {
	ID          int64          `xorm:"pk autoincr"`
	UserID      int64          `xorm:"NOT NULL"`
	IssueID     int64          `xorm:"NOT NULL"`
	ChildID     int64          `xorm:"UNIQUE NOT NULL"`
	CreatedUnix util.TimeStamp `xorm:"created"`
	UpdatedUnix util.TimeStamp `xorm:"updated"`
}

// CreateIssueGroupReaction creates a new group reaction
func CreateIssueGroupReaction(user *User, issue, child *Issue) error {
	sess := x.NewSession()
	defer sess.Close()
	if err := sess.Begin(); err != nil {
		return err
	}

	// Check if it already exists
	exists, err := issueGRExists(sess, issue.ID, child.ID)
	if err != nil {
		return err
	}
	if exists {
		return ErrGroupReactionExists{issue.ID, child.ID}
	}
	// And if it would be circular
	circular, err := issueGRExists(sess, child.ID, issue.ID)
	if err != nil {
		return err
	}
	if circular {
		return ErrCircularGroupReaction{issue.ID, child.ID}
	}

	if _, err := sess.Insert(&IssueGroupReaction{
		UserID:  user.ID,
		IssueID: issue.ID,
		ChildID: child.ID,
	}); err != nil {
		return err
	}

	// Add comment referencing the new dependency
	// if err = createIssueDependencyComment(sess, user, issue, child, true); err != nil {
	// 	return err
	// }

	return sess.Commit()
}

// Check if the group Reaction already exists
func issueGRExists(e Engine, issueID int64, childID int64) (bool, error) {
	return e.Where("(issue_id = ? AND child_id = ?)", issueID, childID).Exist(&IssueGroupReaction{})
}

// RemoveIssueGroupReaction removes a group reaction from an issue
func RemoveIssueGroupReaction(user *User, issue *Issue, child *Issue) (err error) {
	sess := x.NewSession()
	defer sess.Close()
	if err = sess.Begin(); err != nil {
		return err
	}

	issueGRDelete := IssueGroupReaction{IssueID: issue.ID, ChildID: child.ID}

	affected, err := sess.Delete(&issueGRDelete)
	if err != nil {
		return err
	}

	// If we deleted nothing, the dependency did not exist
	if affected <= 0 {
		return ErrGroupReactionNotExists{issue.ID, child.ID}
	}

	// Add comment referencing the removed dependency
	// if err = createIssueDependencyComment(sess, user, issue, child, false); err != nil {
	// 	return err
	// }
	return sess.Commit()
}

func (issue *Issue) GroupReactionChildren() ([]*Issue, error) {
	return issue.getGroupReactionChildren(x)
}

func (issue *Issue) GroupReactionParents() ([]*Issue, error) {
	return issue.getGroupReactionParents(x)
}


// Get children issues
func (issue *Issue) getGroupReactionChildren(e Engine) (issues []*Issue, err error) {
	return issues, e.
		Table("issue_group_reaction").
		Select("issue.*").
		Join("INNER", "issue", "issue.id = issue_group_reaction.child_id").
		Where("issue_id = ?", issue.ID).
		Find(&issues)
}

// Get parents issues
func (issue *Issue) getGroupReactionParents(e Engine) (issues []*Issue, err error) {
	return issues, e.
		Table("issue_group_reaction").
		Select("issue.*").
		Join("INNER", "issue", "issue.id = issue_group_reaction.issue_id").
		Where("child_id = ?", issue.ID).
		Find(&issues)
}


