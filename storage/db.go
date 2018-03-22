package storage

import (
	"database/sql"
	"fmt"
	"reflect"

	_ "github.com/mattn/go-sqlite3"

	"github.com/go-gorp/gorp"
	"github.com/pkg/errors"
)

type DB struct {
	dbMap *gorp.DbMap
}

func DBInit() (*DB, error) {
	fmt.Println("Initializing data store connection")
	db, err := sql.Open("sqlite3", "/tmp/post_db.bin")
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initialize data store connection")
	}

	dbMap := &gorp.DbMap{
		Db:      db,
		Dialect: gorp.SqliteDialect{},
	}

	_ = dbMap.TruncateTables()

	dbMap.AddTableWithName(Card{}, "cardInstances").SetKeys(true, "Id").AddIndex("TrelloCardIndex", "Hash", []string{"TrelloCardId"}).SetUnique(true)
	cardTable, _ := dbMap.TableFor(reflect.TypeOf(Card{}), false)
	cardTable.AddIndex("IssueCardIndex", "BTree", []string{"IssueId", "ListId"})
	dbMap.AddTableWithName(Comment{}, "comments").SetKeys(true, "Id").AddIndex("IssueIdIndex", "Hash", []string{"IssueId"})
	dbMap.AddTableWithName(Issue{}, "issues").SetKeys(true, "Id").AddIndex("IssueIdIndex", "Hash", []string{"IssueId"}).SetUnique(true)

	if err = dbMap.CreateTablesIfNotExists(); err != nil {
		return nil, errors.Wrap(err, "Failed to create data store tables")
	}

	fmt.Println("Data store connection successfully initialized")
	return &DB{
		dbMap: dbMap,
	}, nil
}

func (d *DB) GetOne(holder interface{}, query string, args ...interface{}) error {
	if err := d.dbMap.SelectOne(holder, query, args...); err != nil {
		return err
	}

	return nil
}

func (d *DB) Insert(holders ...interface{}) error {
	if err := d.dbMap.Insert(holders...); err != nil {
		return err
	}
	return nil
}

func (d *DB) Update(holders ...interface{}) (int64, error) {
	count, err := d.dbMap.Update(holders...)
	if err != nil {
		return 0, err
	}
	return count, nil
}
