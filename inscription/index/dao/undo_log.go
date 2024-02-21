package dao

import "github.com/inscription-c/insc/inscription/index/tables"

func (d *DB) FindUndoLog() (undoLogs []*tables.UndoLog, err error) {
	err = d.Order("id desc").Find(&undoLogs).Error
	return
}

func (d *DB) DeleteUndoLog() error {
	s := tables.UndoLog{}
	return d.Exec("delete from " + s.TableName()).Error
}
