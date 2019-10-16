# go randgen

go 版本的 mysql randgen

## go get 安装

```
go get -u  github.com/pingcap/go-randgen/cmd/go-randgen
```

尝试一下：

```
go-randgen -h
```


## 编译安装

 - 安装 go-bindata 命令行工具

```bash
go get -u github.com/jteeuwen/go-bindata/...
```

 - 编译 go-randgen
 
```bash
make all
```

## 特色

 1. 内置一个默认 zz 文件，也就是说你只给一个 yy 文件，就能自动生成 sql
 2. 生成 sql 的过程中可以不连接数据库，非常迅速
 3. 兼容 mysql randgen 的 yy 文件的语法，只要在 yy 文件中没有插入 perl 代码，就可以直接拿过来运行
 4. 和 mysql randgen 支持嵌入 perl 代码类似，go randgen 支持嵌入 lua 代码
 5. 纯 Go 实现，设计得非常灵活，非常易于 Hack
 6. 除了 cmd 包下面的函数，其他包对外暴露的函数的实现全部是无状态，如果需要可以完全当成一个库来调用

## Quick start


### gentest

生成测试 window functions 的 sql：

```bash
# -Y 表示使用的 yy 文件
# -Q 表示生成的查询数量
# -B 表示将数据库构造语句与查询语句分开成两个文件存放
# 这里不需要指定 zz 文件是因为系统自带了一个默认的 zz 文件
./go-randgen gentest -Y examples/windows.yy -Q 10 -B
```

在当前目录下看到`output.data.sql`即是生成的 ddl(表结构定义) 和 dml(初始化表中数据)，
`output.rand.sql`即是根据 yy 文件生成的查询 sql。

上述案例使用的是系统[默认的 zz 文件 ](resource/resource/default.zz.lua)，也可以自己重新写
，然后通过`-Z`参数指定路径，具体规则见语法手册。

如果你不想生成 ddl，只想根据 yy 生成一些 sql，
可以使用`--skip-zz`跳过 ddl 的生成，
不过此时也不允许在 yy 文件中包含表名或者字段相关的关键字。

yy 文件的具体写法也见后面的语法手册

### gendata

根据指定的 zz，往指定的 dsns 中灌入相应数据

```bash
# 在指定的 dsn 中灌入内置 zz 文件定义的数据
# 通过逗号分割多个dsn
./go-randgen gendata --dsns "root:@tcp(127.0.0.1:3306)/randgen,root:@tcp(127.0.0.1:4000)/randgen"
```

### gensql

根据指定的 dsn，解析 yy 文件生成 sql：

```bash
./go-randgen gensql -Y examples/functions.yy \  
             --dsn "root:@tcp(127.0.0.1:3306)/randgen" \ 
             -Q 100
```

注意`gensql`会假设 dsn 中所有表的字段及类型都是一样的（因为 randgen 生成的数据有这个特点）

### exec

指定两个 dsn，直接将生成的 sql 在两个 dsn 上执行，并 dump 出运行结果不一致的 sql

示例：

```bash
./go-randgen exec -Y examples/functions.yy \
             --dsn1 "root:@tcp(127.0.0.1:4000)/randgen" \
             --dsn2 "root:@tcp(127.0.0.1:3306)/randgen" \  
             -Q 100
```

分别在两个 dsn 中先通过内置的 zz 生成数据，然后通过 functions.yy 中定义的规则随机生成 100 条 sql，
在两个 dsn 中同时执行，然后对比执行结果是否一致，如果不一致，
则把相关信息输出到程序执行目录的`dump`目录下
（可以通过`--dump`选项修改 dump 目录）

如果你想让 go-randgen 一直运行下去，而不是执行有限条 sql 后停止，
可以将`-Q`设置为负数，比如`-Q -1`.

注意，默认情况下，对比两个 sql 的执行结果是无序的，比如下面两个运行结果，
go randgen 会认为他们是一样的：

```sql
Result1:

+------+------+
| p    | s    |
+------+------+
|    1 | aaa  |
|    2 | bbb  |
+------+------+

Result2:

+------+------+
| p    | s    |
+------+------+
|    2 | bbb  |
|    1 | aaa  |
+------+------+
```

如果想要精确到 byte 的有序比较的话，可以添加`--order`选项

`exec`也可以通过`--skip-zz`选项跳过数据生成的过程，此时它会采用
类似于`gensql`的方式生成 sql 并执行

### 作为一个库

除了 cmd 目录下的包，其他所有包对外暴露的函数的实现都是无状态的，可以很安全地作为
一个库被反复调用。至于使用的方法，可以参考 cmd 包下相关命令的实现。

## 语法手册

### zz 文件

#### 快速入门

zz 文件是一个 lua 脚本,zz 文件会定义三件事情:

 1. 生成哪些表
 2. 表中哪些字段
 3. 字段中有哪些数据
 
以内置的 zz 文件为例：

```lua
-- 表相关定义
tables = {
    -- 生成的表的记录数
    rows = {10, 20, 30, 90},
    -- 表的字符编码
    charsets = {'utf8', 'latin1', 'binary'},
    -- 表的分区数, 'undef' 表示不分区
    partitions = {4, 6, 'undef'},
}

-- 字段相关定义
fields = {
    -- 需要测试的数据类型
    types = {'bigint', 'float', 'double', 'decimal(40, 20)',
        'char(20)', 'varchar(20)'},
    -- 所有的上面的数字类型都要测试带符合和不带符号两种
    sign = {'signed', 'unsigned'}
}

-- 数据初始化相关定义
data = {
    -- 数字字段的生成方案
    numbers = {'null', 'tinyint', 'smallint',
        '12.991', '1.009', '-9.183',
        'decimal',
    },
    -- 字符串字段的生成方案
    strings = {'null', 'letter', 'english'},
}
```

如上所示，在 zz 文件中必须要有三个 Table 类型的变量，分别是**tables**,
**fields**和**data**.

**tables**中定义表的相关属性，比如示例中的`rows`,`charsets`和`partitions`
（更多的可定义属性见下面的语法手册），这些属性会被求全组合，每种组合生成一张表，
所以示例中的 tables 定义共会生成 4(rows)*3(charsets)*3(partitions)=36 张表

**fields**定义表中的字段信息，这些信息同样会被求全组合，每种组合生成一个字段，
但是上面的示例中生成的字段数目会少于 6(types)*2(sign)=12 个，因为 sign 属性只能
作用在数字类型的字段上面，对于非数字类型，go-randgen 会自动忽略该属性，
所以示例中的配置总计生成的字段数为 4(number)*2(sign)+2(char)=10 个
，注意 randgen 生成的所有表中的字段都是一样的

**data**定义表中的数据，其中 key 代表字段类型
（具体可定义的字段类型见下面的语法手册）
，value 是一个数组，代表该类型字段的
可选值，每生成一条记录，遇到 key 类型的字段时会从 value 随机选择一个作为该条记录的值
，可选值可以是"字面量"或者"生成器"，比如上面示例中对于`numbers`的定义，`null`,
`12.991`等就是字面量，会直接将其作为一个值，而像`tinyint`就是一个生成器，
如果选到它的话，它会从`-128~127`中随机选择一个值生成（具体有哪些生成器见下面的
语法手册）

#### tables

| 字段名称    | 含义    |  可选值  |默认值 |
| --------   | -----   | ---- | ----|
| rows       | 表的记录数  |  任意大于 0 的数字    |[0, 1, 2, 10, 100] |
| charsets   | 字符编码    |  'utf8','utf8mb4','ascii','latin1','binary', 'undef' 表示不显式设置字符集|['undef'] |
| partitions | 分区数      |  任意大于 0 的数字或者 'undef', 'undef' 表示不分区   |['undef'] |

可设置的字段与默认值在源码中见[gendata/tables.go](gendata/tables.go) 的`tablesVars`变量

#### fields

| 字段名称   | 含义    |  可选值  |默认值 |
| --------  | -----   | ---- | ----|
| types     | 字段类型  |任意合法的 mysql 类型|['int', 'varchar', 'date', 'time', 'datetime'] |
| keys      | 索引信息  |'key' 表示加索引,'undef' 表示不加|['undef', 'key']|
| sign      | 是否带符号|'signed', 'unsigned'|['signed']|

可设置的字段与默认值在源码中见[gendata/fields.go](gendata/fields.go) 的`fieldVars`变量

#### data

data 的设置是和 mysql randgen 不太一样的地方，除了支持 numbers
, blobs, temporals, enum, strings 五种梗概类型以外，还可以用
用更细的类型，比如 decimal，bigint 等等，如果存在更细类型的 key 的话，
则以更细类型的定义为准。

比如：

```lua
data = {
    numbers = {'null', 'tinyint', 'smallint',
        '12.991', '1.009', '-9.183',
        'decimal',
    },
    bigint = {100, 10, 3},
}
```

上面这个配置，在遇到 bigint 类型的字段时，每次生成数据会从 100, 10, 3 中随机选择一个，而
不会理会更粗犷的 numbers 的配置。

具体数据类型与梗概数据类型的对应关系见[gendata/data.go](gendata/data.go)
中的`summaryType`变量。

其中 'tinyint', 'smallint', 'decimal' 都是 go randgen 自带的数据生成规则。

go randgen 中支持的所有数据生成规则见
[gendata/generators/register.go](gendata/generators/register.go)
的`init`函数

### yy 文件

#### 快速入门

一个简单的示例：

```yacc

# 单行注释
/*
多行注释
*/

query:
    select
    | select1

select:
    SELECT fields FROM _table
    
fields:
    _field
    | _field_int
```

他的一次生成结果可能如下：

```sql
select1
SELECT 随机的一个字段 FROM 随机的一张表
SELECT 随机的一个整型字段 FROM 随机的一张表
select1
```


#### 注释

 - 单行注释`#`
 - 多行注释`/**/` 

#### 标识符的分类

 - 非终结符：由小写字母,数字或者下划线组成，但是不能以数字开头
 - 终结符：大写字母，特殊字符或者数字组成，但是不能以下划线开头
 - 关键字：下划线开头

> 对于写在表达式右边的非终结符，如果找不到对应的产生式，也会退化成终结符

#### 关键字

关键字都是以下划线开头

获取表名和字段名的接口:

 - `_table`: 从生成的表中随机选择一张
 - `_field`: 从生成的字段中随机选择一个
 - `_field_int`: 从整型字段中随机选择一个
 - `_field_char`: 从 char 和 varchar 类型字段中随机选择一个
 
随机生成数据的一些糖 (字符相关的会在两边自动生成双引号):

 - `_digit`: 随机生成一个 0-9 的数字
 - `_letter`: 随机生成一个 'a' 到 'z' 之间的字母
 - `_english`: 随机生成一个英文单词
 - `_int`: 随机生成一个整型
 - `_date`: 生成`yyyy-MM-dd`格式的随机日期
 - `_year`: 随机生成一个年份
 - `_time`: 随机生成一个`hh:mm:ss`的随机时间
 - `_datetime`: 随机生成一个`yyyy-MM-dd hh:mm:ss`的随机时间


没有写全，代码位于[链接中的 NewKeyfun 方法 ](/gendata/gendata.go)，
可以自行查看

#### 嵌入 lua 代码

可以在大括号（"{}"）的包围中写 lua 代码，调用 print 可以想要的内容拼接到 sql 中

```
query:{a = 1}
    CREATE TABLE 
    {print(string.format("t%d", a))} (a INT)
```

以上代码始终生成 sql 为`CREATE TABLE t1 (a INT)`

在代码块中可以调用 lua 标准库中的任意函数，比如：

```
# 每次随机生成 10-20 的随机数
query:
    {print(math.random(10,20))}
```


正常的代码块会在每次分支被运行到的时候执行一遍。

go randgen 支持在文件的头部插入一个代码块，这个代码块在整个
sql 执行的过程中只会执行一次，称为**头部代码块**，主要用于变量或者函数的申明：

```

# 头部代码块对后面 sql 生成需要的一些变量或函数的进行申明
{
i = 1
a = 100
function add(num1, num2)
    return num1 + num2    
end
}

query:
   select

select:
   SELECT * FROM _table WHERE where_clause
   
where_clause:
   _field_int > {print(i)}
   | _field_char > {print(a)}
   | _field_int + _field_int > {print(add(i, a))}
   
```


通过大括号包围 lua 代码看起来会和 lua 本身的 table 语法相矛盾，
但是你不用担心，我在解析的时候已经作了处理，可以放心大胆地
在代码块中使用 table 语法：

```
{
f={a=1, b=3}
arr={0,2,3,4}
}

query:
  {print(arr[f.a])} | {print(arr[f.b])}
```

上面的代码将只会生成"0"或者"3"（注意 lua 数组的下标是从 1 开始的）

这个示例并没有什么实际意义，只是表达个意思


#### 常用模式

 - 递归地嵌套子查询
 
```
query:
    select

select:
    SELECT * FROM
    (select)
    WHERE _field_int > 10
    | SELECT * FROM _table WHERE _field_char = _english
```

 - 可能为空的规则
 
```
order:
    ASC
    |DESC
    |    # 空规则
    
#....省略其他规则
```

 - 生成多条相邻的 sql 语句
 
有的时候我们希望相关的几条 sql 生成在相邻的位置，比如在测试
Prepared statement 时，下面例子改自[examples/functions.yy](examples/functions.yy) 

```
query:
	SET @stmt = {print('"')} select {print('"')};
	PREPARE stmt FROM @stmt_create ; 
	EXECUTE stmt ;
	
select:
    SELECT * FROM _table
```

此时如果你指定生成的 sql 数量为 3（即`-Q`参数指定为 3）的话，那么就
会生成如下的 sql

```sql
SET @stmt = " SELECT * FROM _table ";
PREPARE stmt FROM @stmt_create; 
EXECUTE stmt;
```

指定生成 6 条 sql 的话，就会把上面的 sql 生成两遍。

假如指定生成 sql 数目为 2 的话，那么会生成：

```sql
SET @stmt = " SELECT * FROM _table ";
PREPARE stmt FROM @stmt_create; 
```

从这里我们也可以看出`;`的语义，这个语义继承自 mysql randgen，
表示一次性生成数条相邻的 sql

 - 测试 create 语句时创建名字不冲突的表


方案一: 插入 lua 脚本,利用头部代码块

```
# 申明 i 为 1
{
i = 1
}

query: 
    create
    
create:
    CREATE TABLE 
    {print(string.format("table%d", i)); i = i+1}
    (a int)
```

生成结果：

```sql
CREATE TABLE table1 (a int);
CREATE TABLE table2 (a int);
CREATE TABLE table3 (a int);
......
```

方案 2：先创建表，然后再把它删除了，利用之前提到的`;`符号

```
query:
    create
    
create:
    CREATE TABLE t (a int); DROP TABLE t
```

生成结果：

```
CREATE TABLE t (a int);
DROP TABLE t;
CREATE TABLE t (a int);
DROP TABLE t;
...
```

## How to hack

相比 mysql randgen，go randgen 最大的特点就是易于 Hack
，几乎没有任何硬编码，当你觉得缺少什么特性时，可以非常
方便地自己加上

### hack zz data

如果你觉得 go randgen 在 zz 文件中 data 字段提供的
数据生成指令不够用时，
可以进入[gendata/generators/register.go](gendata/generators/register.go)
的`init`方法里添加。

假设你在里面添加了一个`aaa`指令，除了能够在
zz 的 data 字段中使用`"aaa"`指令外，
在 yy 文件中也会自动增加一个`_aaa`关键字可以使用

### hack yy key word

如果觉得 yy 中提供的关键字不够用，可以在
[gendata/gendata.go](gendata/gendata.go)
中的`NewKeyfun`方法中添加。



## 与 Mysql randgen 不同的地方

 - 不要在规则尾部加分号, mysql randgen 有这个习惯，但是 go randgen 不需要这么做，我们不依赖这个分号
 来判断不同的规则。当然，你加了也没什么问题，为了兼容 mysql randgen，
 作了额外的处理
 - 数据初始化时使用`insert`，而不是`insert ignore`，在给 unsigned 类型列
生成数据时会自动适配非负数，尝试最多 10 次随机，直到生成正数，如果十次都是负数，
则直接赋予 1
 - zz 文件中 data 的定义可以使用更加精确的数据类型，而不是只有 mysql randgen 中的四种
 - 生成 sql 时可以不连接数据库，利用在生成 ddl 时自动记录下的 schema，非常迅速
