package main

import (
	"os"
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/stretchr/testify/require"
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

func TestGenDataCredentialSources(t *testing.T) {
	defer os.Unsetenv("GO_RANDGEN_DSNS")
	require.NoError(t, os.Setenv("GO_RANDGEN_DSNS", "dsn-one,dsn-two"))

	cmd := newGenDataCmd()
	require.NoError(t, cmd.PreRunE(cmd, nil))
	require.Equal(t, []string{"dsn-one", "dsn-two"}, gendataDsns)
}
