package models

import "code.gitea.io/gitea/modules/util"

// IssueDispatch represents an issue dispatch info
type IssueDispatch struct {
	ID            int64          `xorm:"pk autoincr"`
	UserID        int64          `xorm:"NOT NULL"`
	IssueID       int64          `xorm:"UNIQUE NOT NULL"`
	DispatchedID  int64          `xorm:"UNIQUE NOT NULL"`  // a issue binding with pr
	PullRequestID int64          `xorm:"UNIQUE NOT NULL"`
	CreatedUnix   util.TimeStamp `xorm:"created"`
	UpdatedUnix   util.TimeStamp `xorm:"updated"`
}
