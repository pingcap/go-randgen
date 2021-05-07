tables = {
       -- names => ['A','B','C','D','E','F','G','H','I','J','K','L','M','N','O','P','Q','R','S','T','U','V','W','X','Y','Z', 'AA', 'BB', 'CC', 'DD', 'EE', 'FF', 'GG', 'HH', 'II', 'JJ', 'KK', 'LL', 'MM', 'NN', 'OO', 'PP'],
        -- support 0
        rows = {1, 10, 20, 25, 50, 75, 100},
        charsets = {'utf8', 'latin1', 'binary'},
        partitions = {'undef'},
};

fields = {
    types = {'int', 'tinyint', 'smallint', 'bigint', 'decimal(40, 20)',
     'float', 'double', 'char(20)', 'varchar(20)', 'enum', 'set', 'datetime',
      'bool',  'timestamp', 'year', 'date'},

    sign = {'signed', 'unsigned'},
    keys= {'key','undef'}
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