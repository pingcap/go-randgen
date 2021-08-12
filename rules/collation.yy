# collation 影响 topN
# mpp `Sort` is not supported now
# test point
    # data type -- The string data types are CHAR, VARCHAR, BINARY, VARBINARY, BLOB, TEXT, ENUM, and SET.
    # aggregation function - count,max,min
    # agg,topN,join,sub query
    # string function -- ascii,upper 等
    # 运算符 -- +,like,or,in 等
    # window function todo

# select aggregation(expression1) from table1 join table2 where expression2 group expression3
query:
    select agg_expression from _table as t1 join_type_on _table as t2 on expression
    | select agg_expression from _table as t1 join_type_on _table as t2 on expression having having_exp
    | select agg_expression from _table as t1 where t1. _field_char not in (select _field_char from _table as t2 where expression)
    | select agg_expression from _table as t1 where t1. _field_char in (select _field_char from _table as t2 where expression)
    | select agg_expression from _table as t1 where not exists (select _field_char from _table as t2 where expression)
    | select agg_expression from _table as t1 where exists (select _field_char from _table as t2 where expression)
    | select agg_expression from _table as t1 where t1. _field_char comparison sub_query_modifier (select _field_char from _table as t2 where expression)
    | select agg_expression from _table as t1 where t1. _field_char comparison sub_query_modifier (select _field_char from _table as t2 where expression)
    | select agg_expression from _table as t1 where t1. _field_char comparison (select _field_char from _table as t2 where expression order_by_limit1)
    | select agg_expression from _table as t2 where t2. _field comparison (select agg_expression  from _table as t1 having having_exp)
#    # having
#    | select agg_expression as x from _table as t1 group by t1. _field_char having x not in (select _field_char from _table as t2 where aggregation_condition)
#    | select agg_expression as x from _table as t1 group by t1. _field_char having x not in (select _field_char from _table as t2 where condition)
#    | select agg_expression as x from _table as t1 group by t1. _field_char having not exists (select _field_char from _table as t2 where aggregation_condition)
#    | select agg_expression as x from _table as t1 group by t1. _field_char having exists (select _field_char from _table as t2 where aggregation_condition)
#    | select agg_expression as x from _table as t1 group by t1. _field having x comparison_operation sub_query_modifier (select _field from _table as t2 where condition)
#    | select agg_expression as x from _table as t1 group by t1. _field having x comparison_operation sub_query_modifier (select _field from _table as t2 where aggregation_condition)
#    | select agg_expression as x from _table as t1 group by t1. _field having x comparison_operation (select _field from _table as t2 where condition order_by_limit1)
#    | select agg_expression as x from _table as t2 group by t2. _field having x comparison_operation (select common_aggregation as y from _table as t1)


having_exp:
    agg_expression
    | agg_expression comparison agg_expression
    | agg_expression comparison const_value
    | agg_expression is null
    | agg_expression comparison null
    | agg_expression between 0 and 3
    | agg_expression between "A" and "Y"
    | agg_expression in ("b","y","B",null,"Y","ABC")
    | case agg_expression when null then null end
    | case when null then agg_expression end
    | case when agg_expression then agg_expression end
    | case when agg_expression then agg_expression else agg_expression end
    | ifnull(null,agg_expression)
    | ifnull(agg_expression comparison agg_expression, agg_expression)
    | nullif(null,agg_expression) is null
    | if(null,agg_expression,agg_expression)
    | if( agg_expression,null,null)

join_type_on:
    inner join
    | cross join
    | left join
    | right join

agg_func:
    count(
    | max(
    | min(

agg_operator:
    agg_func
    | agg_func distinct

agg_operator_v:
    t1. _field_char
    | t1. _field_char
    | t1. _field_char
    | const_value
    | ret_num_func
    | ret_str_func
    | if_when_expression
    | expression

agg_expression:
    agg_operator agg_operator_v)
    | count(*)
    | count(distinct agg_operator_v,agg_operator_v)
    | count(distinct agg_operator_v,agg_operator_v,agg_operator_v)

str_expression:
    t1. _field_char is_operator
    | t1. _field_char like_operator
    | t1. _field_char in_operator
    | t1. _field_char between_operator
    | t1. _field_char between_operator
    | ret_num_func comparison t1. _field_int
    | ret_num_func comparison const_int_value
    | ret_str_func comparison t1. _field_char
    | ret_str_func comparison const_str_value
    | t1. _field_char comparison t1. _field_char
    | const_str_value comparison t1. _field_char
    | t1. _field_char comparison t1. _field_char
    | const_str_value comparison t1. _field_char
    | t2. _field_char comparison t1. _field_char
    | const_str_value comparison t2. _field_char
    | t1. _field_char comparison t2. _field_char
    | const_str_value comparison t2. _field_char
    | t1. _field_char comparison t2. _field_char
    | const_str_value comparison t2. _field_char


if_when_expression:
    if(str_expression,func_str_v,func_str_v)
    | case when str_expression then func_str_v end
    | case when str_expression then func_str_v else func_str_v end


expression:
    str_expression
    | str_expression
    | str_expression
    | str_expression
    | str_expression
    | str_expression
    | not (str_expression)
    | (if_when_expression) comparison (if_when_expression)
    | ((if_when_expression) comparison (if_when_expression)) logical_operator str_expression
    | ((if_when_expression) comparison (if_when_expression)) logical_operator ((if_when_expression) comparison (if_when_expression))
    | not ((if_when_expression) comparison (if_when_expression))
    | (str_expression) logical_operator (str_expression)

# not
logical_operator:
    and
    | or
    | xor

comparison:
    =
    | >
    | <
    | <>
    | >=
    | <=
    | !=
#    | <=>

is_operator:
    is is_operator_v
    | is not is_operator_v

is_operator_v:
    null
    | unknown
#    | true
#    | false

# todo strcmp
like_operator:
    like like_operator_v
    | not like like_operator_v
    | rlike like_operator_v
    | not rlike like_operator_v

like_operator_v:
    const_str_value
    | "A%"
    | "a%"
    | "ABC"
    | "A*"
    | "a*"

in_operator:
    in ( in_operator_v )
    | not in ( in_operator_v )

in_operator_v:
    const_str_value
    | t1. _field_char, t1. _field_char, t1. _field_char
    | t1. _field_char, const_str_value
    | t2. _field_char, t2. _field_char, t2. _field_char
    | t2. _field_char, const_str_value
    | t1. _field_char, t2. _field_char, const_str_value
    | t1. _field_char, t2. _field_char, t2. _field_char
    | t2. _field_char, const_str_value, const_str_value, const_str_value, const_str_value, const_str_value

between_operator:
    between between_operator_v and between_operator_v
    | not between between_operator_v and between_operator_v

between_operator_v:
    const_str_value
    | t1. _field_char
    | t2. _field_char

ret_num_func:
    strcmp(func_str_v,func_str_v)
    | find_in_set(func_str_v,func_str_v)
    | instr(func_str_v,func_str_v)
    | locate(func_str_v,func_str_v)
    | char_length(func_str_v)
    | length(func_str_v)

ret_str_func:
    concat(func_str_v,func_str_v)
    | ifnull(func_str_v,func_str_v)
#    | if(strcmp(func_str_v,func_str_v),func_str_v,func_str_v)
#    | case when strcmp(func_str_v,func_str_v) then func_str_v end
#    | case when strcmp(func_str_v,func_str_v) then func_str_v else func_str_v end
    | nullif(func_str_v,func_str_v)
    | case func_str_v when func_str_v then func_str_v end
    | case func_str_v when func_str_v then func_str_v else func_str_v end


func_str_v:
    const_str_value
    | t1. _field_char
    | t2. _field_char

const_value:
    const_str_value
    | "0"
    | "1"
    | "-1"
    | "20"
    | null
    | const_int_value

const_int_value:
    0
    | 1
    | 1
    | 2
    | 2
    | 3
    | -1
    | 10
    | _tinyint

const_str_value:
    _english
    | _english
    | _english
    | _english
    | _english
    | _english
    | _english
    | NULL

order_by_limit1:
    order by pk desc limit 1
    | order by pk asc limit 1
    | order by pk desc limit 0
    | order by pk asc limit 0

sub_query_modifier:
    some
    | any
    | all