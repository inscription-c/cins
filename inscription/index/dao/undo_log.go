package dao

import "github.com/inscription-c/insc/inscription/index/tables"

// FindUndoLog is a method that retrieves all undo logs from the database in descending order of their IDs.
// It returns a slice of pointers to UndoLog structs and an error.
// The method uses the Order method of the DB struct to order the undo logs by their IDs in descending order.
// It then uses the Find method of the DB struct to retrieve the undo logs and assigns them to undoLogs.
// If an error occurs during the retrieval, it returns the undoLogs and the error.
func (d *DB) FindUndoLog() (undoLogs []*tables.UndoLog, err error) {
	err = d.Order("id desc").Find(&undoLogs).Error
	return
}

// DeleteUndoLog is a method that deletes all undo logs from the database.
// It returns an error.
// The method creates a new UndoLog struct and uses its TableName method to get the name of the table.
// It then uses the Exec method of the DB struct to execute a SQL delete statement on the table.
// If an error occurs during the deletion, it returns the error.
func (d *DB) DeleteUndoLog() error {
	s := tables.UndoLog{}
	return d.Exec("delete from " + s.TableName()).Error
}
