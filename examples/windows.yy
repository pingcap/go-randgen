/* test window funtion */

query:
    select

select:
    SELECT
           window_function,
           fieldA,
           fieldB
    FROM (
	SELECT _field AS fieldA, _field AS fieldB
	FROM _table
	ORDER BY LOWER(fieldA), LOWER(fieldB)
    ) as t
    WINDOW window_name AS (window_spec)

window_function:
           CUME_DIST() over_clause AS 'cume_dist'
           | DENSE_RANK() over_clause AS 'dense_rank'
           | FIRST_VALUE(fieldA) null_treatment over_clause AS 'first_value'
           | LAST_VALUE(fieldA) null_treatment over_clause AS 'last_value'
           | LAG(fieldA, lag_para) null_treatment over_clause AS 'lag'
           | LEAD(fieldA, lag_para) null_treatment over_clause AS 'lead'
           | NTH_VALUE(fieldA, number) null_treatment over_clause AS 'nth_value'
           | NTILE(number) over_clause AS 'ntile'
           | PERCENT_RANK() over_clause AS 'percent_rank'
           | RANK() over_clause AS 'rank'
           | ROW_NUMBER() over_clause AS 'row_number'


window_clause:
	window_name AS (window_spec)

window_name:
	w1

window_spec:
	partition_clause order_clause frame_clause
	| partition_clause order_clause frame_clause
	| partition_clause order_clause frame_clause
	| window_name partition_clause order_clause frame_clause

partition_clause:
	PARTITION BY LOWER(fieldB)

order_clause:
	#ORDER BY LOWER(fieldA) order_clause_indication
	#| ORDER BY LOWER(fieldB) order_clause_indication
	ORDER BY LOWER(fieldA) order_clause_indication , LOWER(fieldB) order_clause_indication
	| ORDER BY LOWER(fieldB) order_clause_indication , LOWER(fieldA) order_clause_indication

order_clause_indication:
	ASC | DESC

frame_clause:
	frame_units frame_extent

# TODO: RANGE logical offset based on current row's value
# Ex: RANGE BETWEEN INTERVAL 1 WEEK PRECEDING AND CURRENT ROW
frame_units:
	ROWS | RANGE

frame_extent:
	frame_start | frame_between

frame_between:
	BETWEEN frame_start AND frame_end

frame_start:
  CURRENT ROW
| UNBOUNDED PRECEDING
| number PRECEDING
| UNBOUNDED FOLLOWING  # https://github.com/pingcap/tidb/issues/11002
| number FOLLOWING

frame_end:
  CURRENT ROW
| UNBOUNDED PRECEDING # https://github.com/pingcap/tidb/issues/11001
| number PRECEDING
| UNBOUNDED FOLLOWING
| number FOLLOWING

over_clause:
	OVER (window_spec) | OVER window_name

number:
      -1 | 0 | 1 | 2 | 11 | 91 | 1250951168

null_treatment:
      RESPECT NULLS
| IGNORE NULLS  # https://github.com/pingcap/tidb/issues/10556
|

lag_para:
	number | number , NULL | number , number