/* test deep sub query */

query:
    select

select:
    SELECT * FROM
    (select)
    WHERE _field > 10
    | SELECT * FROM _table WHERE _field > 'a'