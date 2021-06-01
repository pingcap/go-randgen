
query:
     select_all order_by limit 

order_by:
    ORDER BY x1,x2,x3,x4

limit:
    LIMIT 10

select_table:
    select _field_int x1, _field_int x2, _field_char x3, _field x4 from _table

alias_table:
    t1 | t2

alias_name:
    x1 | x2 | x3 | x4

select_all:
      select_table
      | select alias_table . x1, (func) x2 ,  alias_table . x3, alias_table . x4 from (select_all) t1 join (select_all) t2 on t1. alias_name = t2. alias_name
	  | (select_all UNION ALL select_all)

arg:
   t1.x1 | t2.x2 | t1.x2 | t2.x1 | _digit

func:
   arg + arg |
   arg - arg |
   - arg |
   arg * arg |
   arg / arg 

