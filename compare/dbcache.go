package compare

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"sync"
)

// reuse sql.DB object if have the same dsn
type dbCache struct {
	dbs   map[string]*sql.DB
	mutex *sync.RWMutex
}

func (d *dbCache) initDb(dsn string) (*sql.DB, error) {
	key := dbCacheKey(dsn)
	d.mutex.RLock()
	if db, ok := d.dbs[key]; ok {
		d.mutex.RUnlock()
		return db, nil
	}
	d.mutex.RUnlock()

	db, err := OpenDBWithRetry("mysql", dsn)
	if err != nil {
		return nil, err
	}

	d.mutex.Lock()
	d.dbs[key] = db
	d.mutex.Unlock()

	return db, nil
}

func dbCacheKey(dsn string) string {
	sum := sha256.Sum256([]byte(dsn))
	return hex.EncodeToString(sum[:])
}

var cache = &dbCache{
	dbs:   make(map[string]*sql.DB),
	mutex: &sync.RWMutex{},
}
