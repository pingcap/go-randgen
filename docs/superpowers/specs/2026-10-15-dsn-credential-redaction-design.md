# DSN Credential Redaction Design

## Goal

Prevent database credentials from being exposed through standard output, standard error, or persisted comparison logs while preserving enough non-secret connection information for troubleshooting.

## Scope

- Redact credentials from every runtime log message that currently includes a DSN.
- Redact credentials before DSNs are written into comparison dump files.
- Keep the original DSN unchanged when it is passed to `database/sql`.
- Restrict comparison dump directories to mode `0700` and dump files to mode `0600`.
- Cover MySQL DSNs, PostgreSQL URL DSNs, and PostgreSQL keyword/value DSNs.

Changing how DSNs are supplied on the command line is outside this change. Environment-variable or protected-config-file input can be addressed separately.

## Design

Add one shared DSN-redaction helper in the `compare` package so command code and connection retry code use identical behavior.

The helper will preserve non-secret details such as database user, host, port, and database name, while replacing password values with `******`. It will recognize:

- MySQL DSNs such as `user:password@tcp(host:4000)/database`.
- PostgreSQL URLs such as `postgres://user:password@host/database`.
- PostgreSQL keyword/value strings such as `user=user password=password host=host dbname=database`.

If a DSN cannot be parsed or safely recognized, the helper will return `<redacted-dsn>` instead of returning the original input. Empty DSNs remain empty.

Callers will continue to pass the original DSN to `sql.Open`. Only values crossing an output boundary will be redacted.

## Output Boundaries

The following outputs must use the redacted value:

- Connection-open and ping retry messages.
- Fatal connection messages in the `exec`, `gensql`, and `gendata` commands.
- Successful target-database summaries.
- DSN headings stored in comparison dump files.

Database-driver errors will continue to be reported, but tests will ensure the application itself never formats the original DSN into those messages.

## File Permissions

Comparison dump directories will be created with mode `0700`. Individual comparison log files will be created with mode `0600`. Actual permissions may be further restricted by the process umask.

Generated SQL output permissions are not part of this credential-specific change because those files do not contain DSNs.

## Tests

Tests will first demonstrate the current leak, then verify:

- Passwords are removed from supported MySQL and PostgreSQL forms.
- Non-secret connection details remain available.
- Unknown input fails closed and does not return the original value.
- Dump content contains redacted DSNs and never contains plaintext passwords.
- Dump files are created without group or other-user permissions.
- Existing package tests continue to pass.

## Acceptance Criteria

- No plaintext DSN password is emitted by the identified stdout/stderr paths.
- No plaintext DSN password is persisted in comparison logs.
- Database connections still receive the unmodified DSN.
- The full Go test suite passes.
