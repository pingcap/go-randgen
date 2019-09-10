/* test create unique table */

{
i = 1
}

query:
    create

create:
    CREATE TABLE
    {print(string.format("table%d", i)); i = i+1}
    (a int)