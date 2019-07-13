package transaction

import (
	"github.com/jinzhu/gorm"
)

// ITransaction is
type ITransaction interface {
	GetDB(withTransaction bool) interface{} // this is for get the database with two mode with transaction or without transaction
	CommitTransaction(tx interface{})       // commit the transaction
	RollbackTransaction(tx interface{})     // rollback transaction
	RollbackOnException(tx interface{})     // for emergency rollback
}

// GormTransactionDB is
type GormTransactionDB struct {
	DB *gorm.DB
}

// NewGormTransactionDB is
func NewGormTransactionDB(db *gorm.DB) *GormTransactionDB {
	return &GormTransactionDB{db}
}

// GetDB is
func (g *GormTransactionDB) GetDB(withTransaction bool) interface{} {
	if withTransaction {
		return g.DB.Begin()
	}
	return g.DB
}

// CommitTransaction is
func (g *GormTransactionDB) CommitTransaction(tx interface{}) {
	tx.(*gorm.DB).Commit()
}

// RollbackTransaction is
func (g *GormTransactionDB) RollbackTransaction(tx interface{}) {
	tx.(*gorm.DB).Rollback()
}

// RollbackOnException is common handler for rollback the transaction
// to avoid database lock when something goes wrong in transaction state
// use with defer right after we call GetDB(true)
func (g *GormTransactionDB) RollbackOnException(tx interface{}) {
	// catch the error
	if err := recover(); err != nil {

		// rollback it!
		tx.(*gorm.DB).Rollback()

		// repanic so we can get where it happen in log!
		panic(err)
	}
}
