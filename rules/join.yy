query:
    select

select:
    SELECT hint_begin inl_merge_join(t1, t2) */ col_list FROM viewortable  t1, viewortable t2 where condition4 and condition3 last;
    SELECT hint_begin inl_hash_join(t1) */ col_list from viewortable  t1, viewortable t2 where condition4 and condition3 order by t1.pk, t2.pk;
    SELECT hint_begin TIDB_inlj(t1) */ col_list from viewortable  t1, viewortable t2 where condition4 and condition3 order by t1.pk, t2.pk;
    SELECT hint_begin inl_merge_join(t1) */ col_list from viewortable  t1, viewortable t2 where condition4 and condition3 order by t1.pk, t2.pk;
    SELECT hint_begin hash_join(t2) */ col_list from viewortable  t1, viewortable t2 where condition4 and condition3 order by t1.pk, t2.pk;
    SELECT hint_begin merge_join(t2) */ col_list from viewortable  t1, viewortable t2 where condition4 and condition3 order by t1.pk, t2.pk;
    SELECT hint_begin inl_merge_join(t1, t2) */ col_list FROM viewortable  t1 right join viewortable t2 on condition4 where condition3 last;
    SELECT hint_begin inl_hash_join(t1) */ col_list from viewortable  t1 right join viewortable t2 on condition4 where condition3 order by t1.pk, t2.pk;
    SELECT hint_begin TIDB_inlj(t1) */ col_list from viewortable  t1 right join viewortable t2 on condition4 where condition3 order by t1.pk, t2.pk;
    SELECT hint_begin inl_merge_join(t1) */ col_list from viewortable  t1 right join viewortable t2 on condition4 where condition3 order by t1.pk, t2.pk;
    SELECT hint_begin hash_join(t2) */ col_list from viewortable  t1 right join viewortable t2 on condition4 where condition3 order by t1.pk, t2.pk;
    SELECT hint_begin merge_join(t2) */ col_list from viewortable  t1 right join viewortable t2 on condition4 where condition3 order by t1.pk, t2.pk;
    SELECT hint_begin inl_merge_join(t1, t2) */ col_list FROM viewortable  t1 right join viewortable t2 on condition4 where condition3 last;
    SELECT hint_begin inl_hash_join(t1) */ col_list from viewortable  t1 left join viewortable t2 on condition4 where condition3 last;
    SELECT hint_begin TIDB_inlj(t1) */ col_list from viewortable  t1 left join viewortable t2 on condition4 where condition3 order by t1.pk, t2.pk;
    SELECT hint_begin inl_merge_join(t1) */ col_list from viewortable  t1 left join viewortable t2 on condition4 where condition3 order by t1.pk, t2.pk;
    SELECT hint_begin hash_join(t2) */ col_list from viewortable  t1 left join viewortable t2 on condition4 where condition3 order by t1.pk, t2.pk;
    SELECT hint_begin merge_join(t2) */ col_list from viewortable  t1 left join viewortable t2 on condition4 where condition3 order by t1.pk, t2.pk;
    alter table _table add index {print(string.format("t%d", math.random(10,200000000000)))} (_field, _field_char (10));

col_list0:
    max(t1. _field_int), min(t1. _field_int), sum(t1. _field_int), count(t1. _field_int), bit_and(t1. _field_int), bit_or(t1. _field_int), bit_xor(t1. _field_int), round(stddev_samp(t1. _field_int), 4), round(var_samp(t1. _field_int), 4), round(avg(t1. _field_int), 4)

col_list1:
    count(distinct(t1. _field)), count(distinct t1. _field,t1. _field)

col_list2:
    t1.pk, t2.pk, exists (SELECT * from viewortable t where t. _field = t1. _field)

col_list3:
    t1.pk, t2.pk, (SELECT count(*) from _table t3 where t3. _field_int > t2. _field_int)

col_list4:
    count(*)

col_list5:
    t1.pk, t2.pk, case when t1. _field_int < _int then 0 else 1 end

col_list:
    t1.pk, t2.pk

condition1:
    (t2. _field_char in (_english, _english, _english, _english, _letter) and  t1. _field_int  in ({print(math.random(1,100))}, {print(math.random(1,100))}, {print(math.random(1,100))}, {print(math.random(1,100))}, {print(math.random(1,100))}, {print(math.random(1,100))})
    )

condition2:
    t2. _field_int > {print(string.format("%d", math.random(-1000,1000)))} and t1._field_int <= {print(string.format("%d", math.random(-1000,1000)))}

condition3:
    t1. _field_int != {print(string.format("%d", math.random(-1000,1000)))}

condition44:
    t1. _field = t2. _field and t1. _field = t2. _field

condition5:
    t1. _field = t2. _field or t1. _field = t2. _field

condition6:
    t1. _field_int > _int and t1. _field_int <= _int

condition7:
    t1. _field_int = t2. _field

condition4:
    t1. _field_int = t2. _field_int and t1. _field_int < _int

onepartition:
    partition({print(string.format("p%d", math.random(1,4)))})

viewortable:
    {a = math.random(3, 3) if (a == 1) then print("v1") elseif (a == 2) then print("v2") else print("_table") end}

last1:
    group by t1. _field, t2. _field

last:
    order by t1.pk, t2.pk

hint_begin:
    /*+