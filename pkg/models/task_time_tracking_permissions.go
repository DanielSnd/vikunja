package models

import (
	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

func (t *TaskTimeTracking) CanRead(s *xorm.Session, a web.Auth) (bool, int, error) {
	task := Task{ID: t.TaskID}
	return task.CanRead(s, a)
}

func (t *TaskTimeTracking) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	task := Task{ID: t.TaskID}
	return task.CanWrite(s, a)
}

func (t *TaskTimeTracking) canUserModifyTaskTimeTracking(s *xorm.Session, a web.Auth) (bool, error) {
	task := Task{ID: t.TaskID}
	canWrite, err := task.CanWrite(s, a)
	if err != nil || !canWrite {
		return canWrite, err
	}

	saved := &TaskTimeTracking{ID: t.ID, TaskID: t.TaskID}
	if err := getTaskTimeTrackingSimple(s, saved); err != nil {
		return false, err
	}

	if shareAuth, is := a.(*LinkSharing); is {
		return shareAuth.getUserID() == saved.UserID, nil
	}

	return a.GetID() == saved.UserID, nil
}

func (t *TaskTimeTracking) CanUpdate(s *xorm.Session, a web.Auth) (bool, error) {
	return t.canUserModifyTaskTimeTracking(s, a)
}

func (t *TaskTimeTracking) CanDelete(s *xorm.Session, a web.Auth) (bool, error) {
	return t.canUserModifyTaskTimeTracking(s, a)
}
