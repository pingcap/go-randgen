/* test deep sub query */

query:
    select

select:
    SELECT * FROM
    (select)
    WHERE _field_int > 10
    | SELECT * FROM _table WHERE _field_char = _english