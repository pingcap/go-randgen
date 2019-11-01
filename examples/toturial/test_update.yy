/*

test update

used in unittest, if you change it, you should change the relational unittest in `cmd/go-randgen/gentest_test.go`

*/

query:{table = _table()}
    BEGIN ; update ; select ; END

update:
    UPDATE {print(table)} SET _field_int = 10

select:
    SELECT * FROM {print(table)}