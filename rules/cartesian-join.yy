# 以下使用 cartesian join 的场景不一定准确，只是实际测试的使用使用的 cartesian join

# inner / left / right join ，join 条件非等值 / 无条件，如：
# select * from t left join t1 on t.a>t1.a;
# CARTESIAN inner join
# CARTESIAN left outer join
# CARTESIAN right outer join

# not in 关联 / 非关联子查询，关联子查询关联列非等值，如：
# select * from t1 where t1.b not in (select b from t where t.a=10);
# select * from t1 where t1.b not in (select b from t where t1.a=t.a or t1.b>t.b);
# not exists 关联子查询，关联子查询关联列非等值，如：
# select * from t1 where  not exists (select b from t where t1.a>t.a);
# CARTESIAN anti semi join

# exists 关联子查询，关联子查询关联列非等值，如：
# select * from t1 where exists (select b from t where t1.a>t.a);
# CARTESIAN semi join

# 其他 cartesian join 的一些场景
# comparison_oper some/any/all sub_query 关联 / 非关联
# comparison_oper sub_query(order by limit 1) comparison_oper 非等值关联子查询
# comparison_oper sub_query(aggregation return 1 or 0 value) comparison_oper 非等值关联子查询
# select * from t1 where t1.b > some (select b from t);
# select * from t1 where t1.b > (select b from t where t1.a limit 1);
# select * from t1 where t1.b > (select count(*) from t where t1.a);

# 暂不考虑多层嵌套
query:
    sub_query
    | join_query

sub_query:
    # where
    select * from _table as t1 where t1. _field not in (select _field from _table as t2 where condition)
    | select * from _table as t1 where t1. _field not in (select _field from _table as t2 where associate_condition)
    | select * from _table as t1 where not exists (select _field from _table as t2 where associate_condition)
    | select * from _table as t1 where exists (select _field from _table as t2 where associate_condition)
    | select * from _table as t1 where t1. _field comparison_operation sub_query_modifier (select _field from _table as t2 where condition)
    | select * from _table as t1 where t1. _field comparison_operation sub_query_modifier (select _field from _table as t2 where condition)
    | select * from _table as t1 where t1. _field comparison_operation (select _field from _table as t2 where condition order_by_limit1)
    | select * from _table as t2 where t2. _field comparison_operation (select common_aggregation as x from _table as t1 having aggregation_condition)
    # having
    | select common_aggregation as x from _table as t1 group by t1. _field having x not in (select _field from _table as t2 where aggregation_condition)
    | select common_aggregation as x from _table as t1 group by t1. _field having x not in (select _field from _table as t2 where condition)
    | select common_aggregation as x from _table as t1 group by t1. _field having not exists (select _field from _table as t2 where aggregation_condition)
    | select common_aggregation as x from _table as t1 group by t1. _field having exists (select _field from _table as t2 where aggregation_condition)
    | select common_aggregation as x from _table as t1 group by t1. _field having x comparison_operation sub_query_modifier (select _field from _table as t2 where condition)
    | select common_aggregation as x from _table as t1 group by t1. _field having x comparison_operation sub_query_modifier (select _field from _table as t2 where aggregation_condition)
    | select common_aggregation as x from _table as t1 group by t1. _field having x comparison_operation (select _field from _table as t2 where condition order_by_limit1)
    | select common_aggregation as x from _table as t2 group by t2. _field having x comparison_operation (select common_aggregation as y from _table as t1)

join_query:
    select * from _table as t1 join_type _table as t2 on condition
    | select * from _table as t1 join_type _table as t2 on associate_condition

common_aggregation:
    num_aggregation
    | agg_func( _field )
    | agg_func( distinct _field )
    | count(*)

num_aggregation:
    num_agg_func( distinct _field_int)
    | num_agg_func( _field_int)

# 返回值是 num 类型的表达式
num_expression:
    num_field
    | num_field num_operation num_field
    | common_field comparison_operation common_field
    | common_field is bool_value
    | common_field is not bool_value
    | common_field in (common_field,common_field,common_field)
    | common_field not in (common_field,common_field,common_field)
    | common_field between common_field and common_field
    | common_field not between common_field and common_field
    | case num_field when num_field then num_field end
    | case when num_field then num_field end
    | case when num_field then num_field else num_field end
    | ifnull(num_field,num_field)
    | nullif(num_field,num_field)
    | if(num_field,num_field,num_field)
    | common_field like common_value
    | common_field not like common_value
    | common_field rlike common_value
    | common_field not rlike common_value
    | num_func(num_field)

# 关联 expression
associate_expression:
    t1. _field
    | t1. _field_int num_operation num_field
    | t1. _field comparison_operation common_field
    | t1. _field is bool_value
    | t1. _field is not bool_value
    | t1. _field in (common_field,common_field,common_field)
    | t1. _field not in (common_field,common_field,common_field)
    | t1. _field between common_field and common_field
    | t1. _field not between common_field and common_field
    | case t1. _field when common_field then common_field end
    | case when t1. _field then common_field end
    | case when t1. _field then common_field else common_field end
    | ifnull(t1. _field,common_field)
    | nullif(t1. _field,common_field)
    | if(t1. _field,common_field,common_field)
    | t1. _field like common_value
    | t1. _field not like common_value
    | t1. _field rlike common_value
    | t1. _field not rlike common_value
    | num_func(t1. _field_int)

aggregation_expression:
    x
    | x num_operation num_field
    | x comparison_operation common_field
    | x is bool_value
    | x is not bool_value
    | x in (common_field,common_field,common_field)
    | x not in (common_field,common_field,common_field)
    | x between common_field and common_field
    | x not between common_field and common_field
    | case x when common_field then common_field end
    | case when x then common_field end
    | case when x then common_field else common_field end
    | ifnull(x,common_field)
    | nullif(x,common_field)
    | if(x,common_field,common_field)
    | x like common_value
    | x not like common_value
    | x rlike common_value
    | x not rlike common_value
    | num_func(x)

aggregation_condition:
    aggregation_expression
    | aggregation_expression or aggregation_expression
    | aggregation_expression and aggregation_expression
    | aggregation_expression or aggregation_expression and aggregation_expression
    | not aggregation_expression

num_condition:
    num_expression
    | num_expression or num_expression
    | num_expression and num_expression
    | num_expression or num_expression and num_expression
    | not num_expression

condition:
    num_condition

associate_condition:
    associate_expression
    | associate_expression or associate_expression
    | associate_expression and associate_expression
    | associate_expression or associate_expression and associate_expression
    | not associate_expression


# 一些不常用的聚合函数暂不考虑
agg_func:
    min
    | max
    | count

num_agg_func:
    avg
    | sum

hint_name:
    inl_merge_join
    | inl_hash_join
    | hash_join
    | merge_join
    | inl_join

join_type:
    inner join
    | left join
    | right join

col_list:
    count(*)


# 精度考虑，去掉一些函数，如 acos，sqrt
num_func:
    abs
    | ceil
    | ceiling
    | crc32
    | floor
    | round
    | sign
    | isnull

# Error 1690: BIGINT UNSIGNED value is out of range
num_operation:
    /
#    | +
#    | -
#    | *
    | %
#    | >>
#    | <<
    | <=>
    | ^
    | comparison_operation

# Error 1105: We don't support <=> all or <=> any now
comparison_operation:
    =
    | >
    | <
    | <>
    | >=
    | <=
    | !=

# 增加 _field 的概率
common_field:
    t2. _field
    | t2. _field
    | t2. _field
    | t2. _field
    | common_value


# 增加 _field_int 的概率
num_field:
    t2. _field_int
    | t2. _field_int
    | t2. _field_int
    | t2. _field_int
    | num_value

common_value:
    _letter
    | _english
    | _date
    | _time
    | _datetime
    | "b"
    | "y"
    | "x"
    | "null"
    | num_value

num_value:
    _year
    | _digit
    | null
    | 12.991
    | 1.009
    | -9.183
    | 0
    | -1
    | 1
    | 12
    | 13
    | "1"
    | ""
    | "0"

bool_value:
   TRUE | FALSE | UNKNOWN | NULL

sub_query_modifier:
    some
    | any
    | all

order_by_limit1:
    order by pk desc limit 1
    | order by pk asc limit 1