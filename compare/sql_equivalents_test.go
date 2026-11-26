package compare

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLSafetyRejectsPrivilegedEquivalents(t *testing.T) {
	for _, statement := range []string{
		"SELECT * FROM mysql.user",
		"SELECT * FROM `mysql` . /* catalog */ `user`",
		"SELECT * FROM pg_authid",
		"SELECT * FROM pg_catalog.\"pg_authid\"",
		"SELECT * FROM pg_auth_members",
		"SELECT * FROM pg_shadow",
		"UPDATE mysql.user", "DELETE FROM pg_authid",
		"INSERT INTO pg_catalog.pg_auth_members", "REPLACE INTO mysql.global_priv",
		"WITH data AS (SELECT * FROM pg_authid) SELECT * FROM data",
		"SELECT set_config", "SELECT pg_ls_logdir", "SELECT pg_ls_tmpdir",
		"SELECT pg_ls_archive_statusdir", "SELECT pg_sleep",
		"SELECT pg_sleep_for", "SELECT pg_sleep_until",
		"SELECT query_to_xml('SELECT 1', false, false, '')",
		"SELECT query_to_xmlschema('SELECT 1', false, false, '')",
		"SELECT query_to_xml_and_xmlschema('SELECT 1', false, false, '')",
		"SELECT table_to_xml('sample', false, false, '')",
		"SELECT table_to_xmlschema('sample', false, false, '')",
		"SELECT table_to_xml_and_xmlschema('sample', false, false, '')",
		"SELECT schema_to_xml('sample', false, false, '')",
		"SELECT schema_to_xmlschema('sample', false, false, '')",
		"SELECT schema_to_xml_and_xmlschema('sample', false, false, '')",
		"SELECT database_to_xml(false, false, '')",
		"SELECT database_to_xmlschema(false, false, '')",
		"SELECT database_to_xml_and_xmlschema(false, false, '')",
		"SELECT cursor_to_xml('sample', 1, false, false, '')",
		"SELECT cursor_to_xmlschema('sample', false, false, '')",
		"SELECT pg_catalog.QuErY_To_XmL('SELECT 1', false, false, '')",
		`SELECT "pg_catalog"."table_to_xml"('sample', false, false, '')`,
		`SELECT "query_to_xml"($sql$SELECT 1$sql$, false, false, '')`,
		"CREATE TABLE sample (id INT) ENGINE=FEDERATED",
		"CREATE TABLE sample (id INT) ENGINE='FEDERATED'",
		"ALTER TABLE sample ENGINE=\"FEDERATED\"",
		"CREATE TABLE sample (id INT) CONNECTION='placeholder'",
	} {
		t.Run(statement, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			assert.Error(t, ValidateSQLSafety(statement))
			_, err = query(db, statement)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "blocked by safety policy")
			_, err = exec(db, statement)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "blocked by safety policy")
			_, err = ExecSqlsInDbs([]string{statement}, db)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "blocked by safety policy")
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestSQLSafetyEquivalentCompatibility(t *testing.T) {
	for _, statement := range []string{
		"SELECT * FROM app.users",
		"SELECT mysql FROM sample",
		"SELECT * FROM pg_catalog.pg_tables",
		"SELECT * FROM pg_catalog.pg_roles",
		"CREATE TABLE sample (id INT) ENGINE=InnoDB",
		"CREATE TABLE sample (id INT) ENGINE='InnoDB'",
		"SELECT engine = 'InnoDB' FROM sample",
		"SELECT ENGINE '",
		"SELECT ENGINE ''",
		"SELECT 'pg_authid, mysql.user, set_config, FEDERATED CONNECTION'",
		"SELECT 1 /* pg_authid, mysql.user, set_config, FEDERATED CONNECTION */",
		"SELECT pg_authid_count, set_config_status FROM sample",
		"SELECT 'query_to_xml, table_to_xml, schema_to_xml, database_to_xml, cursor_to_xml'",
		"SELECT $$query_to_xmlschema, table_to_xml_and_xmlschema$$",
		"SELECT 1 /* query_to_xml, database_to_xmlschema */",
		"SELECT query_to_xml_result FROM sample",
		"SELECT xml_is_well_formed('<sample/>')",
		"SELECT xmlconcat('<a/>'::xml, '<b/>'::xml)",
		"SELECT xpath('/sample', '<sample/>'::xml)",
		"SELECT xmlserialize(DOCUMENT xmlparse(DOCUMENT '<sample/>') AS text)",
	} {
		assert.NoError(t, ValidateSQLSafety(statement), statement)
	}
}
