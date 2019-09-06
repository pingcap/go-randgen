tables = {
    rows = {10, 20, 30, 40, 50, 60, 70, 80, 90},
    -- SHOW CHARACTER SET;
    charsets = {'utf8', 'utf8mb4', 'ascii', 'latin1', 'binary'},
    -- partition number
    partitions = {4, 6, 8, 15, 'undef'},
}

fields = {
    types = {'bigint', 'float', 'double', 'decimal(40, 20)',
        'char(20)', 'varchar(20)'},
    sign = {'signed', 'unsigned'}
}

data = {
    numbers = {'null', 'tinyint', 'smallint',
        '12.991', '1.009', '-9.183',
        'decimal',
    },
    strings = {'null', 'letter', 'english'},
}