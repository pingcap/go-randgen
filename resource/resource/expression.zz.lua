tables = {
    rows = {1, 10, 20, 30, 90},
    -- SHOW CHARACTER SET;
    charsets = {'utf8'},
    -- partition number
    partitions = {'undef'},
}

fields = {
    types = {'int', 'tinyint', 'smallint', 'bigint', 'decimal(40, 20)',
     'float', 'double', 'char(20)', 'varchar(20)', 'enum', 'set', 'datetime',
      'bool',  'timestamp', 'year', 'date'},

    sign = {'signed', 'unsigned'},
    keys= {'key'}
}

data = {
    numbers = {'null', 'tinyint',
        '12.991', '1.009', '-9.183','0','-1','1'
    },
    enum={'"y"','"b"','1','"x"','"null"'},
     bool = {1, 0, null},
       year = {'null', 'year'},
       datetime = {'null', 'datetime'},
       timestamp = {'null', 'datetime'},
       date = {'null', 'date'},
       strings = {'null', 'letter', 'english','1','0'},
}
