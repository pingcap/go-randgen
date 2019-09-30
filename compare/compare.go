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

	for _, sql := range sqls {
		if sql == "" {
			continue
		}

		consistent, dsn1Res, dsn2Res := BySql(sql, db1, db2, nonOrder)

		if !consistent {
			if err := visitor(sql, dsn1Res, dsn2Res); err != nil {
				return err
			}
		}
	}

	return nil
}

func BySql(sql string, db1 *sql.DB, db2 *sql.DB, nonOrder bool) (consistent bool, dsn1Res *DsnRes,
	dsn2Res *DsnRes) {
	r1, err1 := query(db1, sql)
	r2, err2 := query(db2, sql)

	if err1 == driver.ErrBadConn {
		log.Printf("Error: connection to dsn1 error, %v \n", err1)
	}

	if err2 == driver.ErrBadConn {
		log.Printf("Error: connection to dsn2 error, %v \n", err2)
	}

	dsn1Res = &DsnRes{r1, err1}
	dsn2Res = &DsnRes{r2, err2}

	if !errConsistent(err1, err2) {
		return false, dsn1Res, dsn2Res
	}

	// err all not nil, think it is consistent without need to compare
	if err1 != nil && err2 != nil {
		return true, dsn1Res, dsn2Res
	}

	// compare
	if nonOrder {
		if !r1.NonOrderEqualTo(r2) {
			return false, dsn1Res, dsn2Res
		}
	} else {
		if !r1.BytesEqualTo(r2) {
			return false, dsn1Res, dsn2Res
		}
	}

	return true, dsn1Res, dsn2Res
}

func errConsistent(err1 error, err2 error) bool {
	return (err1 == nil && err2 == nil) || (err1 != nil && err2 != nil)
}
