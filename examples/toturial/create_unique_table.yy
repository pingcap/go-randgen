/*

test create unique table

used in unittest, if you change it, you should change the relational unittest in `cmd/go-randgen/gentest_test.go`

*/

{
i = 1
}

query:
    create

create:
    CREATE TABLE
    {print(string.format("table%d", i)); i = i+1}
    (a int)