query:
  query_type

query_type:
  simple_select | mixed_select | aggregate_select

mixed_select:
   SELECT distinct straight_join select_list FROM join

simple_select:
   SELECT distinct straight_join simple_select_list FROM join

aggregate_select:
   SELECT distinct straight_join aggregate_select_list FROM join

distinct: DISTINCT | | | |

straight_join:  | | | | | | | | | | | STRAIGHT_JOIN

select_list:
	new_select_item |
	new_select_item , select_list |
        new_select_item , select_list

simple_select_list:
     nonaggregate_select_item
     | simple_select_list

aggregate_select_list:
     aggregate_select_item
     | aggregate_select_list

new_select_item:
    nonaggregate_select_item
	| aggregate_select_item

nonaggregate_select_item:
        t1. _field_int

aggregate_select_item:
        aggregate t1. _field_int )

aggregate:
	COUNT( distinct | SUM( distinct | MIN( distinct | MAX( distinct

left_right:
	LEFT
	| RIGHT

join:
    _table as t1 left_right outer JOIN _table as t2 ON join_condition

join_condition:
   int_condition | char_condition

int_condition:
    t1. _field_int = t2. _field_int

char_condition:
    t1. _field_char = t2. _field_char
