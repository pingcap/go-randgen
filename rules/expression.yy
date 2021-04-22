query:
    select

select:{a = _field_list()}
    SELECT hint_begin use_index_merge(t1,{print(string.gsub(a,"`pk`,",""))}) */ prjcommon FROM _table  t1 where condition6;

hint_begin:
    /*+

prjcommon:
    *

conditiontrue:
    1

prj_pk:
    t1.pk, t2.pk

# bug https://github.com/pingcap/tics/issues/1522
condition5:
    t1. _field =1  or t1. _field !=null

#bug https://github.com/pingcap/tics/issues/1523
condition51:
    if(t1. _field,null,not t1. _field)

# bug https://github.com/pingcap/tics/issues/1537 https://github.com/pingcap/tics/issues/1536
condition52:
    ifnull(t1. _field,null)

# bug https://github.com/pingcap/tics/issues/1538
condition53:
    case t1. _field when null then null end
    | case t1. _field when t1. _field_int / t1. _field_int, t1. _field then null end
    | case when null then t1. _field end
    | case when t1. _field then t1. _field end
    | nullif(null,t1. _field,null) is null
    | ifnull(null,t1. _field)
    | ifnull(t1. _field_int / t1. _field_int, t1. _field)

# bug https://github.com/pingcap/tics/issues/1540
condition54:
   t1. _field_int / t1. _field_int or (t1. _field_int * t1. _field_int)
    | t1. _field_int / t1. _field_int or (t1. _field_int - t1. _field_int)
    | t1. _field_int / t1. _field_int or (t1. _field_int + t1. _field_int)

# bug https://github.com/pingcap/tics/issues/1543 https://github.com/pingcap/tics/issues/1540
# in 的问题较多，需要修复后再次验证
condition7:
    t1. _field in (1,1,2,2,3,3)
    | t1. _field in (null,1,1,0,0)
    | t1. _field in ('-9.183','0','-1','1','0')
    | (t1. _field_char in (_english, _english, _english, _english, _letter) or  t1. _field_int  in ({print(math.random(1,100))}, {print(math.random(1,100))}, {print(math.random(1,100))}, {print(math.random(1,100))}, {print(math.random(1,100))}, {print(math.random(1,100))}))
    | t1. _field in ("y","y",0,0)
    | t1. _field in ("y","y",0,0,null,null)

prj7:
    concat(t1. _field,null, t1. _field)
    | concat(t1. _field,t1. _field)
    | concat(t1. _field_int ,t1. _field_int)
    | concat(null,null)

condition6:
    | not t1. _field between null and 200
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
    | strcmp(t1. _field,t1. _field)
    | strcmp(t1. _field,null)
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


condition8:
    not t1. _field_int / t1. _field_int
    | t1. _field_int / t1. _field_int or not (t1. _field_int & t1. _field_int)
    | t1. _field_int / t1. _field_int or not (t1. _field_int > t1. _field_int)
    | t1. _field_int / t1. _field_int or not (t1. _field_int >> t1. _field_int)
    | t1. _field_int / t1. _field_int or not (t1. _field_int < t1. _field_int)
    | t1. _field_int / t1. _field_int or not (t1. _field_int << t1. _field_int)
    | t1. _field_int / t1. _field_int or not (t1. _field_int <> t1. _field_int)
    | t1. _field_int / t1. _field_int or not (t1. _field_int <=> t1. _field_int)
    | t1. _field_int / t1. _field_int or not (t1. _field_int != t1. _field_int)
    | t1. _field_int / t1. _field_int or not (t1. _field_int ^ t1. _field_int)
    | t1. _field_int / t1. _field_int or not (t1. _field_int % t1. _field_int)
    | t1. _field_int / t1. _field_int or not (t1. _field_int != t1. _field_int)


onepartition:
    partition({print(string.format("p%d", math.random(1,4)))})

viewortable:
    {a = math.random(3, 3) if (a == 1) then print("v1") elseif (a == 2) then print("v2") else print("_table") end}

last1:
    group by t1. _field, t2. _field

last:
    order by t1.pk, t2.pk

