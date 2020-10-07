package kvs

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type KVS struct {
	DBPath string
	Bucket string
	db     *bolt.DB
}

func Open(dbPath string, bucket string) (kvs *KVS, err error) {
	kvs = &KVS{
		DBPath: dbPath,
		Bucket: bucket,
	}
	kvs.db, err = bolt.Open(dbPath, 0666, nil)
	if err != nil {
		return
	}
	return
}

func (kvs *KVS) Close() {
	kvs.db.Close()
}

func (kvs *KVS) Get(key string) (json []byte) {
	tx, err := kvs.db.Begin(false)
	if err != nil {
		panic(fmt.Sprintf("Cannot start transaction in database '%s'", kvs.DBPath))
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != bolt.ErrTxClosed {
			panic(fmt.Sprintf("Cannot rollback database '%s' with error: %v", kvs.DBPath, err))
		}
	}()

	b := tx.Bucket([]byte(kvs.Bucket))
	if b == nil {
		return
	}
	json = b.Get([]byte(key))

	return
}

func (kvs *KVS) Set(key string, json []byte) (err error) {
	tx, err := kvs.db.Begin(true)
	if err != nil {
		panic(fmt.Sprintf("Cannot start transaction in database '%s'", kvs.DBPath))
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != bolt.ErrTxClosed {
			panic(fmt.Sprintf("Cannot rollback database '%s' with error: %v", kvs.DBPath, err))
		}
	}()

	b, err := tx.CreateBucketIfNotExists([]byte(kvs.Bucket))
	if err != nil {
		return
	}

	err = b.Put([]byte(key), json)
	if err != nil {
		return err
	}

	// commit
	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}
