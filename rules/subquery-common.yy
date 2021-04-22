query:
    select

num_agg_func_distinct_para1:
    avg(
    | sum(
    | avg( distinct
    | sum( distinct

agg_func_distinct_para1:
    min(
    | max(
    | count(
    | min( distinct
    | max( distinct
    | count( distinct

agg_func:
    num_agg_func_distinct_para1
    | agg_func_distinct_para1

subquery_operator:
    ANY
    | SOME
    | ALL

subquery_operator2:
    EXISTS
    | NOT EXISTS

order_by:
    order by pk desc
    | order by pk asc

scalar_q:
    SELECT _field from _table as t1 where condition  order_by  limit 1
    | SELECT num_agg_func_distinct_para1 t1. _field_int ) from _table as t1 where condition
    | SELECT agg_func_distinct_para1 t1. _field ) from _table as t1 where condition


column_q:
    SELECT _field from _table as t1 where condition


field_t2_random:
    t2. _field_int
    | t2. _field_char
    | null

# correlated_subquery
# derived_table
select:
    SELECT col_list from _table as t2 where field_t2_random operation ( scalar_q )
    | SELECT col_list from _table as t2 where field_t2_random comparison_operation subquery_operator ( column_q)
    | SELECT col_list from _table as t2 where field_t2_random in ( column_q)
    | SELECT col_list from _table as t2 where subquery_operator2 ( column_q )


join_type_where:
     natural join
     | natural left join
     | natural right join


col_list:
    *
    | agg_func agg_field )
    | hint_begin hash_agg() */ agg_func agg_field )
    | hint_begin stream_agg() */ agg_func agg_field )

agg_field:
    t2. _field_int

hint_begin:
    /*+

comparison_operation:
    =
    | >
    | <
    | <>
    | >=
    | <=
    | !=

operation:
#    like
#    | +
#    | -
#    | *
#    | /
#    | %
#    | >>
#    | <<
#    | <=>
#    | ^
    comparison_operation

condition:
    condition_join_column
    | condition_common
    | condition_common and condition_common
    | condition_common or condition_common
    | condition_join_column or condition_common
    | condition_join_column and condition_common
    | not condition_join_column
    | condition_join_column or not condition_common
    | condition_join_column and not condition_common


condition_join_column:
    t1. _field_int operation t2. _field_int
    | t1. _field_char operation t2. _field_char

condition_common:
    condition_null
#    | common_func
    | condition_between
    | condition_in
    | field_random operation value_random

condition_null:
    case t1. _field when null then null end
#    | case t1. _field when t1. _field_int / t1. _field_int then null end
    | case when null then t1. _field end
    | case when t1. _field then t1. _field end
    | case when t1. _field then t1. _field else t1. _field end
    | case when common_func then condition_between else t1. _field end
    | case when condition_between then condition_between end
    | ifnull(null,t1. _field)
#    | ifnull(t1. _field_int / t1. _field_int, t1. _field)
    | nullif(null,t1. _field) is null
    | if(null,t1. _field,t1. _field)
    | if(t1. _field,null,null)
    | if(t1. _field,t1. _field,null)
    | t1. _field is null
    | t1. _field operation null

# t1. _field in (1,1,2,2,3,3) year bug
condition_in:
    t1. _field in (null,2001,2001,2000,2000)
    | t1. _field in ('-9.183','0','-1','1','0')
    | (t1. _field_char in (_english, _english, _english, _english, _letter) or  t1. _field_int  in ({print(math.random(1,100))}, {print(math.random(1,100))}, {print(math.random(1,100))}, {print(math.random(1,100))}, {print(math.random(1,100))}, {print(math.random(1,100))}))
    | t1. _field in ("y","y",0,0)
    | t1. _field in ("y","y",0,0,null,null)
    | t1. _field in (t1. _field,t1. _field)
    | t1. _field in (t1. _field,t1. _field,null)

condition_between:
    not t1. _field between null and 200
    | not t1. _field between t1. _field and null
    | not null between t1. _field and null
    | not null between t1. _field and t1. _field
    | not 0 between t1. _field and 3
    | not t1. _field between 0.111 and t1. _field
    | not t1. _field between 0.111 and 1102221324
    | not t1. _field between 0 and 10
    | not t1. _field between "0" and "y"
    | t1. _field between 10 and 0
    | t1. _field between 0 and 0
    | t1. _field between "0" and "2"

field_random:
    t1. _field
    | null

value_random:
    12.991
    | 1.009
    | -9.183
    | 0
    | -1
    | 1
    | 12
    | 13
    | "y"
    | "b"
    | "x"
    | _year
    | _date
    | _datetime
    | _letter
    | _english
    | "%b%"
    | "%y"
    | "0%"
    | "%1"
    | "%-"
    | null
    | t1. _field



str_func:
    ascii(field_random)
    | bin(field_random)
    | bit_length(field_random)
    | char(field_random)
    | char_length(field_random)
    | character_length(field_random)
    | concat(field_random,field_random)
    | concat_ws(field_random,field_random,field_random)
    | hex(field_random)
    | lcase(field_random)
    | length(field_random)
    | field_random like value_random
    | field_random not like value_random
    | lower(field_random)
    | ltrim(field_random)
    | oct(field_random)
    | octet_length(field_random)
    | quote(field_random)
    | repeat(field_random,0)
    | reverse(field_random)
    | field_random rlike value_random
    | rtrim(field_random)
    | space(field_random)
    | strcmp(field_random,value_random)
    | to_base64(field_random)
    | trim(field_random)
    | ucase(field_random)
    | unhex(field_random)
    | upper(field_random)

num_func:
    field_random operation field_random
    | field_random operation value_random
    | abs(field_random)
    | acos(field_random)
    | asin(field_random)
    | atan(field_random)
    | ceil(field_random)
    | ceiling(field_random)
    | cos(field_random)
    | cot(field_random)
    | crc32(field_random)
    | floor(field_random)
    | round(field_random)
    | sign(field_random)
    | sin(field_random)
    | sqrt(field_random)
    | tan(field_random)

common_func:
    str_func
    | num_func
    | least(t1. _field,t1. _field,t1. _field,t1. _field)
    | least(t1. _field,t1. _field,null)
    | greatest(t1. _field,t1. _field,t1. _field,t1. _field)
    | greatest(t1. _field,t1. _field,null)
    | coalesce(t1. _field,t1. _field,t1. _field,t1. _field)
    | coalesce(null,t1. _field,t1. _field,null)
    | coalesce(null,null)
    | interval(t1. _field,t1. _field,t1. _field,t1. _field)
    | interval(null,t1. _field,t1. _field,t1. _field)
    | interval(t1. _field,t1. _field,null,t1. _field)
    | interval(t1. _field,"y",t1. _field,t1. _field)
    | interval(null,null)
    | isnull(t1. _field)
    | isnull(t1. _field / t1. _field)

