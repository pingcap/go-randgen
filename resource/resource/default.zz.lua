tables = {
    rows = {10, 20, 30, 40, 50, 60, 70, 80, 90},
    -- SHOW CHARACTER SET;
    charsets = {'utf8', 'utf8mb4', 'ascii', 'latin1', 'binary'},
    partitions = {4, 6, 8, 15, 'undef'},
}

fields = {
    types = {'bigint', 'float', 'double', 'decimal(40, 20)',
        'decimal(10, 4)', 'decimal(6, 3)', 'char(20)', 'varchar(20)'},
    sign = {'signed', 'unsigned'}
}

data = {
    numbers = {'null', 'tinyint', 'smallint',
        '12.991', '1.009', '-9.1823',
        '-111.1212', '12.98731', '1.098781',
        '0.112345', '-0.987103', '-0.000000001',
        '0.00000001', '0.999999999', '-0.999999999',
        'decimal',
    },
    strings = {'null', 'letter', 'english', 'string(15)'}
}