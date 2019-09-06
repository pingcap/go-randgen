# go randgen

go port from mysql randgen

目前只支持`_table`和`_field`两个关键字，分别是随即取一张表和随机取一个字段，
暂时不支持插入代码块。

```
make all
./go-randgen -Y examples/windows.yy -Q 10 -B
# 将ddl dml放入数据库中执行以生成数据
mysql -h localhost -P 3306 -u root -p123456 -D test < output.data.sql > mysql.out
```

在当前目录下看到`output.data.sql`即是生成的ddl和dml，
`output.rand.sql`即是根据yy文件生成的查询

使用的是系统默认的zz文件，位于`resource/resource/default.zz.lua`

和randgen细节上的不同：
 1. 关键字（`_`）以及非终结符的左右务必留空格
 2. 对于unsigned类型的变量, 框架会进行10次重试直到生成正值, 如果10次重试都没能生成正值,直接赋值1
 3. 支持单行注释#，以及多行注释/* ... */
 4. 全部以`query`作为根规则（mysql randgen在生成第一条语句时会使用`query_init`作为根规则）


还需优化的地方：

 1. enum字段类型的支持
 2. 根据varchar的长度自动生成在该长度以内的字符串
 3. 支持插入lua代码块
 
 
 
注释（在go randgen里目前全部将其视为普通注释，不予其功能含义）：
 
 1. '#'后面是注释
 2. '/* ... */' 注释将被输出到query log中
 3. '/*! ... */'  只发送给Mysql的sql语句
 4. '/*+JavaDB:Postgres: ..*/'  只给指定的数据库发送该语句 
 

