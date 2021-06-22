package compare

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

// OpenDBWithRetry opens a database specified by its database driver name and a
// driver-specific Data source name. And it will do some retries if the connection fails.
// variablize for mock conveniently
var OpenDBWithRetry = func(driverName, dataSourceName string) (mdb *sql.DB, err error) {
	startTime := time.Now()
	sleepTime := time.Millisecond * 500
	retryCnt := 60
	// The max retry interval is 30 s.
	for i := 0; i < retryCnt; i++ {
		mdb, err = sql.Open(driverName, dataSourceName)
		if err != nil {
			fmt.Printf("open db `%s` failed, retry count %d err %v\n",
				dataSourceName, i, err)
			time.Sleep(sleepTime)
			continue
		}
		err = mdb.Ping()
		if err == nil {
			break
		}
		log.Printf("Error: ping db `%s` failed, retry count %d err %v\n",
			dataSourceName, i, err)
		mdb.Close()
		time.Sleep(sleepTime)
	}
	if err != nil {
		log.Printf("Error: open db `%s` failed %v, take time %v\n",
			dataSourceName, err, time.Since(startTime))
		return nil, err
	}

	return
}

// sql result present one query result receive from database
type SqlResult struct {
	// [row][col][content(nil when it is NULL)]
	Data        [][][]byte
	Rows        map[string]bool
	Header      []string
	ColumnTypes []*sql.ColumnType
	err         error
}

func (s *SqlResult) Contains(row string) bool {
	_, ok := s.Rows[row]
	return ok
}

func (s *SqlResult) NonOrderEqualTo(another *SqlResult) bool {
	if len(s.Rows) != len(another.Rows) {
		return false
	}

	for row := range another.Rows {
		if !s.Contains(row) {
			return false
		}
	}

	return true
}

// if s is equal to another, will return true
func (s *SqlResult) BytesEqualTo(another *SqlResult) bool {
	if len(s.Data) != len(another.Data) {
		return false
	}

	for i := range s.Data {
		if !s.RowBytesEqualTo(another, i, another.Data[i]) {
			return false
		}
	}

	return true
}

func (s *SqlResult) RowBytesEqualTo(another *SqlResult, r int, row [][]byte) bool {
	row1 := s.Data[r]
	row2 := row
	if len(row1) != len(row2) {
		return false
	}

	for i := range row1 {
		if !s.ColBytesEqualTo(another, r, i, row[i]) {
			return false
		}
	}

	return true
}

func (s *SqlResult) ColBytesEqualTo(another *SqlResult, r, c int, col []byte) bool {
	col1 := s.Data[r][c]
	col2 := col
	// all NULL
	if col1 == nil && col2 == nil {
		return true
	}

	if len(col1) != len(col2) {
		return false
	}

	return bytes.Equal(col1, col2)
}

func (result *SqlResult) String() string {
	if result == nil || result.Data == nil || result.Header == nil {
		return "no result"
	}

	// Calculate the max column length
	var colLength []int
	for _, c := range result.Header {
		colLength = append(colLength, len(c))
	}
	for _, row := range result.Data {
		for n, col := range row {
			if l := len(col); colLength[n] < l {
				colLength[n] = l
			}
			if colLength[n] < 4 {
				// for NULL
				colLength[n] = 4
			}
		}
	}

	// The total length
	var total = len(result.Header) - 1
	for index := range colLength {
		colLength[index] += 2 // Value will wrap with space
		total += colLength[index]
	}

	var lines []string
	var push = func(line string) {
		lines = append(lines, line)
	}

	// Write table header
	var header string
	for index, col := range result.Header {
		length := colLength[index]
		padding := length - 1 - len(col)
		if index == 0 {
			header += "|"
		}
		header += " " + col + strings.Repeat(" ", padding) + "|"
	}
	splitLine := "+" + strings.Repeat("-", total) + "+"
	push(splitLine)
	push(header)
	push(splitLine)

	// Write rows data
	for _, row := range result.Data {
		var line string
		for index, col := range row {
			if col == nil {
				col = []byte("NULL")
			}
			length := colLength[index]
			padding := length - 1 - len(col)
			if index == 0 {
				line += "|"
			}
			line += " " + string(col) + strings.Repeat(" ", padding) + "|"
		}
		push(line)
	}
	push(splitLine)
	return strings.Join(lines, "\n")
}

func query(db *sql.DB, sql string) (*SqlResult, error) {
	rows, err := db.Query(sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}

	var allRows [][][]byte
	rowSet := make(map[string]bool)
	for rows.Next() {
		if rows.Err() != nil {
			return nil, err
		}

		var columns = make([][]byte, len(cols))
		var pointer = make([]interface{}, len(cols))
		rowStrBuf := &bytes.Buffer{}
		for i := range columns {
			pointer[i] = &columns[i]
		}
		err := rows.Scan(pointer...)
		if err != nil {
			return nil, err
		}
		for _, colByte := range columns {
			if colByte == nil {
				rowStrBuf.WriteString("NULL\t")
				continue
			}
			rowStrBuf.WriteString(string(colByte) + "\t")
		}

		rowSet[rowStrBuf.String()] = true
		allRows = append(allRows, columns)
	}

	return &SqlResult{Data: allRows, Rows: rowSet, Header: cols, ColumnTypes: types}, nil
}

func exec(db *sql.DB, sql string) (int64, error) {
	result, err := db.Exec(sql)
	if err != nil {
		return 0, nil
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

var execSqls = map[string]bool{
	"delete": true,
	"create": true,
	"update": true,
}

func isExec(sql string) bool {
	if len(sql) <= 6 {
		return false
	}

	prefix := sql[0:6]
	_, ok := execSqls[strings.ToLower(prefix)]
	return ok
}

type SqlExecErr struct {
	sql string
	err error
}

func ExecSqlsInDbs(sqls []string, dbs ...*sql.DB) (string, error) {
	wg := &sync.WaitGroup{}
	wg.Add(len(dbs))

	errCh := make(chan *SqlExecErr, 1)
	c, cancel := context.WithCancel(context.Background())

	for _, db := range dbs {
		go func(db *sql.DB) {
			defer wg.Done()
			for _, sqlStr := range sqls {
				select {
				case <-c.Done():
					break
				default:
					if _, err := db.Exec(sqlStr); err != nil {
						cancel()
						select {
						case errCh <- &SqlExecErr{sqlStr, err}:
						default:
						}
						break
					}
				}
			}
		}(db)
	}

	wg.Wait()
	close(errCh)
	sqlErr := <-errCh
	if sqlErr != nil {
		return sqlErr.sql, sqlErr.err
	}

	return "", nil
}
