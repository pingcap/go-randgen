# go randgen

go版本的mysql randgen

## 如何编译
---

 - 安装go-bindata命令行工具

```bash
go get -u github.com/jteeuwen/go-bindata/...
```

 - 编译go-randgen
 
```bash
make all
```

## Quick start

randgen根据用户指定的zz文件初始化数据库（建表与插入数据），yy文件来随机生成sql。

生成测试window functions的sql：

```bash
# -Y表示使用的yy文件
# -Q表示生成的查询数量
# -B表示将数据库构造语句与查询语句分开成两个文件存放
# 这里不需要指定zz文件是因为系统自带了一个默认的zz文件
./go-randgen -Y examples/windows.yy -Q 10 -B
```

在当前目录下看到`output.data.sql`即是生成的ddl(表结构定义)和dml(初始化表中数据)，
`output.rand.sql`即是根据yy文件生成的查询sql。

上述案例使用的是系统[默认的zz文件](resource/resource/default.zz.lua)，也可以自己重新写
，然后通过`-Z`参数指定路径，具体规则见语法手册。

和randgen细节上的不同：
 3. 支持单行注释#，以及多行注释/* ... */
 4. 全部以`query`作为根规则（mysql randgen在生成第一条语句时会使用`query_init`作为根规则）

## 语法手册
---

### zz文件

zz文件定义数据库的初始化状态，其本身是一个lua脚本文件，以默认的zz文件为例：

```lua
-- 表相关定义
tables = {
    -- 生成的表的记录数
    rows = {10, 20, 30, 90},
    -- 表的字符编码
    charsets = {'utf8', 'latin1', 'binary'},
    -- 表的分区数, 'undef'表示不分区
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

#### tables

| 字段名称    | 含义    |  可选值  |默认值 |
| --------   | -----   | ---- | ----|
| rows       | 表的记录数  |  任意大于0的数字    |[0, 1, 2, 10, 100] |
| charsets   | 字符编码    |  'utf8','utf8mb4','ascii','latin1','binary', 'undef'表示不显式设置字符集|['undef'] |
| partitions | 分区数      |  任意大于0的数字或者'undef', 'undef'表示不分区   |['undef'] |

#### fields

| 字段名称   | 含义    |  可选值  |默认值 |
| --------  | -----   | ---- | ----|
| types     | 字段类型  |任意合法的mysql类型|['int', 'varchar', 'date', 'time', 'datetime'] |
| keys      | 索引信息  |'key'表示加索引,'undef'表示不加|['undef', 'key']|
| sign      | 是否带符号|'signed', 'unsigned'|['signed']|

#### data

data的设置是和mysql randgen不太一样的地方，除了支持numbers
, blobs, temporals, enum, strings五种梗概类型以外，还可以用
用更细的类型，比如decimal，bigint等等，如果存在更细类型的key的话，
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

上面这个配置，在遇到bigint类型的字段时，每次生成数据会从100, 10, 3中随机选择一个，而
不会理会更粗犷的numbers的配置。

其中'null', 'tinyint', 'smallint', 'decimal'都是go randgen自带的数据生成规则。

go randgen中支持的所有数据生成规则见[链接](gendata/generators/register.go)

日后再整理文档

### yy文件

#### 注释

 - 单行注释`#`
 - 多行注释`/**/` 

#### 关键字

获取表名和字段名的接口:

 - `_table`: 从生成的表中随机选择一张
 - `_field`: 从生成的字段中随机选择一个
 - `_field_int`: 从整型字段中随机选择一个
 - `_field_char`: 从char和varchar类型字段中随机选择一个
 
随机生成数据的一些糖:

 - `_digit`: 随机生成一个0-9的数字
 - `_letter`: 随机生成一个'a'到'z'之间的字母
 - `_english`: 随机生成一个英文单词
 - `_int`: 随机生成一个整型
 - `_date`: 生成`yyyy-MM-dd`格式的随机日期
 - `_year`: 随机生成一个年份
 - `_time`: 随机生成一个`hh:mm:ss`的随机时间
 - `_datetime`: 随机生成一个`yyyy-MM-dd hh:mm:ss`的随机时间


没有写全，代码位于[链接](/gendata/gendata.go)的底部，可以自行查看，日后有空再补全

#### 嵌入lua代码

可以在大括号（"{}"）的包围中写lua代码，调用print可以想要的内容拼接到sql中

```
query:{a = 1}
    CREATE TABLE 
    {print(string.format("t%d", a))} (a INT)
```

以上代码始终生成sql为`CREATE TABLE t1 (a INT)`

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

 - 测试create语句时创建名字不冲突的表（插入lua脚本）

```
query: 
    {if(i==nil) then i = 1 end}  # 如果变量还未定义,申明i为1
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

## 与Mysql randgen不同的地方

 - 千万不要在规则尾部加分号, mysql randgen有这个习惯，但是go randgen不要这么做，我们不依赖这个分号
 来判断不同的规则
 - 数据初始化时使用`insert`，而不是`insert ignore`，在给unsigned类型列
生成数据时会自动适配非负数，尝试最多10次随机，直到发现负数，如果十次都是负数，则直接赋予1
 - zz文件中data的定义可以使用更加精确的数据类型，而不是只有mysql randgen中的四种


## 未来开发计划

 - 根据varchar字段的长度自动生成在该长度以内的字符串，避免insert报错，彻底杜绝insert ignore的使用