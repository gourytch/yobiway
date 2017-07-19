package client

import (
	"github.com/boltdb/bolt"
	"log"
	"time"
)

const DATABASE_FNAME = "yobiway.bolt"

var bucketCACHE = []byte("CACHE")

var boltdb *bolt.DB = nil

func BoltDB_init() error {
	var err error
	boltdb, err = bolt.Open(DATABASE_FNAME, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		log.Printf("database %v open error: %s", DATABASE_FNAME, err)
		return err
	}
	boltdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketCACHE)
		if b == nil {
			log.Printf("create bucket %s", bucketCACHE)
			_, err := tx.CreateBucket(bucketCACHE)
			if err != nil {
				log.Fatalf("bucket creation error: %s", err)
				return err
			}
		} else {
			//log.Printf("bucket %s exists", bucketCACHE)
		}
		return err
	})
	return err
}

func BoltDB_close() {
	boltdb.Close()
}

func cache_get(url []byte) (body []byte) {
	boltdb.View(func(tx *bolt.Tx) error {
		body = []byte(tx.Bucket(bucketCACHE).Get(url))
		if body == nil {
			//log.Printf("+ %s not in cache", url)
		} else {
			//log.Printf("+ %s get cached %d bytes", url, len(body))
		}
		return nil
	})
	return
}

func cache_put(url, body []byte) {
	//log.Printf("+ %s <- %d bytes", url, len(body))
	boltdb.Update(func(tx *bolt.Tx) error {
		//log.Printf("... + %s <- %d bytes", url, len(body))
		err := tx.Bucket(bucketCACHE).Put(url, body)
		if err != nil {
			log.Printf("! not cached %s due error: %s", url, err)
		} else {
			//log.Printf("+ %s cached", url)
		}
		return err
	})
}
