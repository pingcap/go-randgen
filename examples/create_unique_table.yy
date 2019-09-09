/* test create unique table */

query:
    {if(i==nil) then i = 1 end} create

create:
    CREATE TABLE
    {print(string.format("table%d", i)); i = i+1}
    (a int)