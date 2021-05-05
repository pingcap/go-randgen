## AST root
query:
    select ;


## _field, _table are go_randgen keywords
select:
    SELECT field FROM _table where ;

field:
    _field |
    * ;

## reverse the value
/* T -> F, F -> T, NULL -> NULL */
bi_logic:
      |     # empty
    NOT ;


## numeric
op_num:
    + |
    - |
    * |
    / ;

## bitwise
op_bit:
    &  |
    {print("|")} |
    ^  |
    << |
    << ;


boolean_case:
    IS TRUE |
    IS FALSE |
    IS NULL ;


## requires explicit cast
where:
    WHERE  bi_logic ( ( CAST( predicate AS BOOLEAN) ) boolean_case ) ;


## _digit is a go_randgen keyword
predicate:
    _field op_num _digit |
    _digit op_bit _field ;