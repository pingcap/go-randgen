/* innerjoin sub query and update sub query
./go-randgen gentest -Y examples/subquery_test.yy --maxrecur 30
*/

query:
    BEGIN ; update ; ROLLBACK | select

update:
    # update sub query
    UPDATE _table
      SET _field_int = _tinyint
    WHERE _field_char in (select_char)

select:
    select_all
    | select_char

select_all:
    SELECT * FROM (select) inner_join
    where
    | SELECT * FROM _table

select_char:
    SELECT _field_char FROM (select_char) inner_join
    WHERE _field_char in (select_char)
    | SELECT a._field_char FROM (SELECT * FROM _table) a
    | SELECT _field_char FROM _table

where:
    WHERE _field_char in (select_char)

inner_join:
    # inner join sub query
    INNER JOIN (select) ON _field_char = _field_char
    |