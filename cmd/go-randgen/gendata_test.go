package main

import (
	"github.com/magiconair/properties/assert"
	"testing"
)

/*func monkeyOpenDB(test func()) {
	oldOpenDB := compare.OpenDBWithRetry
	compare.OpenDBWithRetry = func(driverName, dataSourceName string) (mdb *sql.DB, err error) {
		db, mock, err := sqlmock.New()
		mock.ExpectExec("*")

		return db, err
	}

	test()

	compare.OpenDBWithRetry = oldOpenDB
}*/

func TestGenData(t *testing.T) {
	t.Run("test empty dsn", func(t *testing.T) {
		reInitCmd()
		_, err := executeCommand(rootCmd, "gendata")
		assert.Equal(t, "At least one dsn are required", err.Error())
	})
}
