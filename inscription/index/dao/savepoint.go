package dao

import (
	"errors"
	"github.com/inscription-c/insc/inscription/index/tables"
	"gorm.io/gorm"
)

// ListSavepoint is a method that retrieves a list of all savepoints from the database.
// It returns a slice of pointers to SavePoint structs and an error.
// The method uses the Find method of the DB struct to retrieve the savepoints and assigns them to the list.
// If an error occurs during the retrieval, it checks if the error is a ErrRecordNotFound error from the gorm package.
// If it is, it assigns nil to the error, effectively ignoring the error.
// The method then returns the list of savepoints and the error.
func (d *DB) ListSavepoint() (list []*tables.SavePoint, err error) {
	err = d.Find(&list).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = nil
	}
	return
}

// OldestSavepoint is a method that retrieves the oldest savepoint from the database.
// It returns a pointer to a SavePoint struct and an error.
// The method uses the First method of the DB struct to retrieve the oldest savepoint and assigns it to savepoint.
// If an error occurs during the retrieval, it returns the savepoint and the error.
func (d *DB) OldestSavepoint() (savepoint *tables.SavePoint, err error) {
	err = d.First(&savepoint).Error
	return
}

// DeleteSavepoint is a method that deletes all savepoints from the database.
// It returns an error.
// The method creates a new SavePoint struct and uses its TableName method to get the name of the table.
// It then uses the Exec method of the DB struct to execute a SQL delete statement on the table.
// If an error occurs during the deletion, it returns the error.
func (d *DB) DeleteSavepoint() error {
	s := tables.SavePoint{}
	return d.Exec("delete from " + s.TableName()).Error
}
