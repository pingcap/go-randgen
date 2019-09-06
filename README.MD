# go randgen

go port from mysql randgen

developing now

一些额外的特性：
 1. 对于unsigned类型的变量, 框架会进行10次重试直到生成正值, 如果10次重试都没能生成正值,直接赋值1


还需优化的地方：

 1. enum字段类型的支持
 2. 根据varchar的长度自动生成在该长度以内的字符串
 3. 记录alter table语句对表字段的修改
 
 
 yy文件解析需要注意的地方：
 
 
注释（在go randgen里目前全部将其视为普通注释，不予其功能含义）：
 
 1. '#'后面是注释
 2. '/* ... */' 注释将被输出到query log中
 3. '/*! ... */'  只发送给Mysql的sql语句
 4. '/*+JavaDB:Postgres: ..*/'  只给指定的数据库发送该语句 
 

根规则(go randgen全部以query作为根规则)：

 1. 第一条sql以query_init作为根规则
 2. 剩下的都以query为根规则

yy规则：

 1. 小写是非终结符
 2. 大写和带引号的都是普通字符串
 
 
谢腾进的代码需要修改的地方：

 1. 注释的解析
 2. 错误处理，返回error，而不是fatal
 3. 插入lua代码块
 4. 需要的token:
     1. NonTerminal  全部由小写字母或者下划线组成
     2. Terminal  不以下划线开头，不是全由小写字母或者下划线组成，也可以字符串字面量
     3. Keyword   下划线开头