package dao

import (
	"database/sql"
	"github.com/inscription-c/cins/inscription/index/tables"
)

// AddUndoLog is a method that adds an undo log to the database.
// It takes two parameters: height of type uint32 and sql of type string.
// The height parameter represents the height of the undo log.
// The sql parameter represents the SQL statement of the undo log.
//
// The method first calls the OldestSavepoint method to get the oldest savepoint.
// If an error occurs during this operation, it returns the error.
// If the ID of the oldest savepoint is 0, it returns nil.
//
// The method then creates a new UndoLog struct with the provided height and SQL statement,
// and calls the Create method to add the undo log to the database.
// If an error occurs during this operation, it returns the error.
//
// The method returns an error if any operation fails, or nil if the operations succeed.
func (d *DB) AddUndoLog(height uint32, sql string) error {
	oldest, err := d.OldestSavepoint()
	if err != nil {
		return err
	}
	if oldest.Id == 0 {
		return nil
	}
	return d.Create(&tables.UndoLog{
		Height: height,
		Sql:    sql,
	}).Error
}

// FindUndoLog is a method that retrieves all undo logs from the database in descending order of their IDs.
// It returns a slice of pointers to UndoLog structs and an error.
// The method uses the Order method of the DB struct to order the undo logs by their IDs in descending order.
// It then uses the Find method of the DB struct to retrieve the undo logs and assigns them to undoLogs.
// If an error occurs during the retrieval, it returns the undoLogs and the error.
func (d *DB) FindUndoLog() (*sql.Rows, error) {
	return d.Model(&tables.UndoLog{}).Order("id desc").Rows()
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
