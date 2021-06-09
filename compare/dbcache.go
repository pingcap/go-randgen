package compare

import (
	"database/sql"
	"sync"
)

// reuse sql.DB object if have the same dsn
type dbCache struct {
	dbs   map[string]*sql.DB
	mutex *sync.RWMutex
}

func (d *dbCache) initDb(dsn string) (*sql.DB, error) {
	d.mutex.RLock()
	if db, ok := d.dbs[dsn]; ok {
		d.mutex.RUnlock()
		return db, nil
	}
	d.mutex.RUnlock()

	db, err := OpenDBWithRetry("mysql", dsn)
	if err != nil {
		return nil, err
	}

	d.mutex.Lock()
	d.dbs[dsn] = db
	d.mutex.Unlock()

	return db, nil
}

var cache = &dbCache{
	dbs:   make(map[string]*sql.DB),
	mutex: &sync.RWMutex{},
}
