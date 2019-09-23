package compare

import (
	"database/sql"
	"database/sql/driver"
	"log"
)

type DsnRes struct {
	Res *SqlResult
	Err error
}

type Visitor func(sql string, dsn1Res *DsnRes, dsn2Res *DsnRes) error

func ByDsn(sqls []string, dsn1 string, dsn2 string, nonOrder bool, visitor Visitor) error {

	db1, err := cache.initDb(dsn1)
	if err != nil {
		return err
	}

	db2, err := cache.initDb(dsn2)
	if err != nil {
		return err
	}

	return ByDb(sqls, db1, db2, nonOrder, visitor)
}

func ByDb(sqls []string, db1 *sql.DB, db2 *sql.DB, nonOrder bool, visitor Visitor) error {

	exec := executor{
		visitor:  visitor,
		db1:      db1,
		db2:      db2,
		nonOrder: nonOrder,
	}

	for _, sql := range sqls {
		if sql == "" {
			continue
		}

		if err := exec.executeAndDiff(sql); err != nil {
			return err
		}
	}

	return nil
}

type executor struct {
	visitor  Visitor
	nonOrder bool
	db1      *sql.DB
	db2      *sql.DB
}

func errConsistent(err1 error, err2 error) bool {
	return (err1 == nil && err2 == nil) || (err1 != nil && err2 != nil)
}

// return if result is consistent, true represent consistent
func (e *executor) executeAndDiff(sql string) error {
	r1, err1 := query(e.db1, sql)
	r2, err2 := query(e.db2, sql)

	if err1 == driver.ErrBadConn {
		log.Printf("Error: connection to dsn1 error, %v \n", err1)
	}

	if err2 == driver.ErrBadConn {
		log.Printf("Error: connection to dsn2 error, %v \n", err2)
	}

	dsn1Res := &DsnRes{r1, err1}
	dsn2Res := &DsnRes{r2, err2}

	if !errConsistent(err1, err2) {
		// err not all nil, or all not nil
		if err := e.visitor(sql, dsn1Res, dsn2Res); err != nil {
			return err
		}
		return nil
	}

	// err all not nil, think it is consistent without need to compare
	if err1 != nil && err2 != nil {
		return nil
	}

	// compare
	if e.nonOrder {
		if !r1.NonOrderEqualTo(r2) {
			if err := e.visitor(sql, dsn1Res, dsn2Res); err != nil {
				return err
			}
		}
	} else {
		if !r1.BytesEqualTo(r2) {
			if err := e.visitor(sql, dsn1Res, dsn2Res); err != nil {
				return err
			}
		}
	}

	return nil
}
