tables = {
    rows = {1, 10, 20, 30, 90},
    -- SHOW CHARACTER SET;
    charsets = {'utf8'},
    -- partition number
    partitions = {'undef'},
}

fields = {
    types = {'int','bigint', 'float', 'double', 'decimal(40, 20)',
        'char(20)', 'varchar(20)', 'enum'},
    sign = {'signed', 'unsigned'},
    keys= {'key'}
}

data = {
    numbers = {'null', 'tinyint',
        '12.991', '1.009', '-9.183','0','-1','1'
    },
    enum={'"y"','"b"','1','"x"','"null"'},
    strings = {'"abc"', '"hello"', '"big"','"gg"','"dd"','"ee"','1','0','null'},
}
