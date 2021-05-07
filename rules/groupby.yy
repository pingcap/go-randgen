field_num:
    _field_int
    | null

field_ret1:
    field_num
    | _field
    | _field_char
# tidb panic bug
#    | _time
#    | _datetime

field_ret1_m:
    field_ret1
    | _field_list
    | _field_int_list
    | _field_char_list

field_retm_all:
    field_ret1_m
    | *

num_agg_func_distinct_para1:
    avg
    | sum

agg_func_distinct_para1:
    min
    | max
    | count

agg_func_para1:
    agg_func_distinct_para1
# bug known
#    | bit_and
#    | bit_or
#    | bit_xor
# 精度不一样
#    | std
#    | stddev
#    | stddev_pop
#    | stddev_samp
#    | var_pop
#    | var_samp
#    | variance


group_concat1:
    group_concat(field_ret1_m)
    | group_concat(field_ret1_m order by field_ret1_m asc separator _english)
    | group_concat(field_ret1_m order by field_ret1_m desc separator _english)

# 精度
func_num:
    abs(
#    | acos(
#    | asin(
#    | atan(
    | ceil(
    | ceiling(
#    | cos(
#    | cot(
    | crc32(
    | floor(
    | round(
    | sign(
#    | sin(
#    | sqrt(
#    | tan(
    | isnull(

value_random_num:
    12.991
    | 1.009
    | -9.183
    | 0
    | -1
    | 1
    | 12
    | 13
    | "1"
    | "0"
    | _year
    | null

comparison_operation:
    =
    | >
    | <
    | <>
    | >=
    | <=
    | !=
    | <=>
#    | /
    | %

# DB::TiFlashException: CastStringAsReal is not supported
condition_null:
    case _field when null then null end
    | case when null then _field end
    | case when _field then _field end
    | case when _field then _field else _field end
    | ifnull(null,_field_int)
    | ifnull(_field_int comparison_operation _field_int, _field_int)
    | nullif(null,_field_int) is null
    | if(null,_field_int,_field_int)
    | if( _field_int,null,null)
    | if( _field_int ,_field_int,null)
    | _field_int is null
    | _field_int comparison_operation null

condition_in:
    _field in (null,2001,2001,2000,2000)
    | _field in ('-9.183','0','-1','1','0')
    | _field in ("y","y",0,0)
    | _field in ("y","y",0,0,null,null)
    | _field in (_field,_field)
    | _field in ( _field, _field,null)
    | _field not in(value_random_num,"y","z")

condition_between:
    _field between null and value_random_num
    | _field between _field and null
    | _field between value_random_num and _field
    | _field between value_random_num and value_random_num
    | null between _field and null
    | null between _field and _field
    | value_random_num between _field and value_random_num
    | _field between "0" and "y"

num_selection_expr:
# 概率增加
    field_num
    | field_num
    | field_num
    | field_num
    | field_num comparison_operation field_num
    | field_num comparison_operation value_random_num
    | func_num field_num )
    | condition_between
    | condition_in
    | condition_between
#    | condition_null

num_selection:
    num_selection_expr
    | not num_selection_expr
    | num_selection_expr or num_selection_expr
    | num_selection_expr and num_selection_expr

num_selection_d:
    num_selection
    | distinct num_selection


having_exp:
    agg_selection
    | agg_selection comparison_operation agg_selection
    | agg_selection comparison_operation value_random_num
    | func_num agg_selection )
    | agg_selection is null
    | agg_selection comparison_operation null
    | agg_selection between 0 and 3
    | agg_selection in ("z","y",0)
    | agg_selection in ("y",2002,null)
    | case agg_selection when null then null end
    | case when null then agg_selection end
    | case when agg_selection then agg_selection end
    | case when agg_selection then agg_selection else agg_selection end
    | ifnull(null,agg_selection)
    | ifnull(agg_selection comparison_operation agg_selection, agg_selection)
    | nullif(null,agg_selection) is null
    | if(null,agg_selection,agg_selection)
    | if( agg_selection,null,null)

having_exp_l:
    having_exp
    | not having_exp
    | having_exp and having_exp
    | having_exp or having_exp

agg_selection:
    num_agg_func_distinct_para1(num_selection_d)
    | agg_func_distinct_para1(field_ret1)
    | agg_func_distinct_para1(distinct field_ret1)
    | agg_func_para1(field_ret1)
    | agg_func_para1(num_selection_d)
    | count(*)

cmm_agg_selection:
    agg_func_distinct_para1(field_ret1)
    | agg_func_distinct_para1(distinct field_ret1)
    | agg_func_distinct_para1(field_ret1)
    | agg_func_distinct_para1(distinct field_ret1)
    | agg_func_distinct_para1(field_ret1)
    | agg_func_distinct_para1(distinct field_ret1)
    | agg_func_distinct_para1( num_selection_d)
    | agg_func_para1(field_ret1)
    | agg_func_para1(num_selection_d)
    | count(*)

hint_func:
    stream_agg() */
    | hash_agg() */
    | USE_TOJA(true) */
    | USE_TOJA(false) */
    | AGG_TO_COP() */

# some query may error

simple_group_by:
# sum avg
    | select num_agg_func_distinct_para1( num_selection_d ) from _table as t1
    | select num_agg_func_distinct_para1( num_selection_d ) from _table as t1 group by t1. _field
    | select num_agg_func_distinct_para1( num_selection_d ) from _table as t1 having having_exp_l
    | select num_agg_func_distinct_para1( num_selection_d ) from _table as t1 group by t1. _field having having_exp_l
# count max min
    | select cmm_agg_selection from _table as t1
    | select cmm_agg_selection from _table group by field_ret1
    | select cmm_agg_selection from _table having having_exp_l
    | select cmm_agg_selection from _table group by field_ret1 having having_exp_l

# sum avg
    | select hint_begin hint_func num_agg_func_distinct_para1( num_selection_d ) from _table as t1
    | select hint_begin hint_func num_agg_func_distinct_para1( num_selection_d ) from _table as t1 group by t1. _field
    | select hint_begin hint_func num_agg_func_distinct_para1( num_selection_d ) from _table as t1 having having_exp_l
    | select hint_begin hint_func num_agg_func_distinct_para1( num_selection_d ) from _table as t1 group by t1. _field having having_exp_l
# count max min
    | select hint_begin hint_func cmm_agg_selection from _table as t1
    | select hint_begin hint_func cmm_agg_selection from _table group by field_ret1
    | select hint_begin hint_func cmm_agg_selection from _table having having_exp_l
    | select hint_begin hint_func cmm_agg_selection from _table group by field_ret1 having having_exp_l

#    | select group_concat1 from _table
#    | select group_concat1 from _table group by field_ret1_m


# example sql aggregation
# select sum(distinct a),avg(a) from t group by b having count(a) >1;
# select sum(distinct a+b) from t group by a;
# hash_agg stream_agg
query:
    variable_set; simple_group_by

variable_set:
    set @@tidb_allow_batch_cop=0
    | set @@tidb_allow_batch_cop=1
    | set @@tidb_allow_batch_cop=2
    | set @@tidb_allow_mpp=on
    | set @@tidb_allow_mpp=off
    | set @@tidb_broadcast_join_threshold_count=0
    | set @@tidb_broadcast_join_threshold_count=1
    | set @@tidb_broadcast_join_threshold_count=10240
    | set @@tidb_broadcast_join_threshold_size=0
    | set @@tidb_broadcast_join_threshold_size=1
    | set @@tidb_broadcast_join_threshold_size=104857600
    | set @@tidb_distsql_scan_concurrency=1
    | set @@tidb_distsql_scan_concurrency=15
    | set @@tidb_enable_cascades_planner=on
    | set @@tidb_enable_cascades_planner=off
    | set @@tidb_enable_chunk_rpc=on
    | set @@tidb_enable_chunk_rpc=off
    | set @@tidb_enable_index_merge=off
    | set @@tidb_enable_index_merge=on
    | set @@tidb_enable_parallel_apply=on
    | set @@tidb_enable_parallel_apply=off
    | set @@tidb_enable_vectorized_expression=on
    | set @@tidb_enable_vectorized_expression=off
    | set @@tidb_executor_concurrency=1
    | set @@tidb_executor_concurrency=5
    | set @@tidb_index_lookup_size=10
    | set @@tidb_index_lookup_size=1
    | set @@tidb_index_lookup_size=2000
    | set @@tidb_index_serial_scan_concurrency=1
    | set @@tidb_index_serial_scan_concurrency=10
    | set @@tidb_init_chunk_size=1
    | set @@tidb_init_chunk_size=3
    | set @@tidb_init_chunk_size=32
    | set @@tidb_max_chunk_size=32
    | set @@tidb_max_chunk_size=1024
    | set @@tidb_opt_agg_push_down=off
    | set @@tidb_opt_agg_push_down=on
    | set @@tidb_opt_correlation_exp_factor=0
    | set @@tidb_opt_correlation_exp_factor=1
    | set @@tidb_opt_correlation_threshold=0.9
    | set @@tidb_opt_correlation_threshold=0.5
    | set @@tidb_opt_distinct_agg_push_down=on
    | set @@tidb_opt_distinct_agg_push_down=off
    | set @@tidb_opt_insubq_to_join_and_agg=on
    | set @@tidb_opt_insubq_to_join_and_agg=off
    | set @@tidb_opt_prefer_range_scan=0
    | set @@tidb_opt_prefer_range_scan=1




hint_begin:
    /*+