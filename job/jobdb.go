package job

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"
	"time"

	"github.com/boltdb/bolt"
)

var (
	jobBucket = []byte("jobs")
)

func GetDB(path string) *BoltJobDB {
	if path != "" && !strings.HasSuffix(path, "/") {
		path += "/"
	}
	path += "jobdb.db"
	database, err := bolt.Open(path, 0600, &bolt.Options{Timeout: time.Second * 10})
	if err != nil {
		log.Fatal(err)
	}
	return &BoltJobDB{
		path:   path,
		dbConn: database,
	}
}

type JobDB interface {
	GetAll() ([]*Job, error)
	Get(id string) (*Job, error)
	Delete(id string)
	Save(job *Job) error
	Close()
}

type BoltJobDB struct {
	dbConn *bolt.DB
	path   string
}

func (db *BoltJobDB) Close() {
	db.dbConn.Close()
}

func (db *BoltJobDB) GetAll() ([]*Job, error) {
	allJobs := []*Job{}

	err := db.dbConn.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(jobBucket)
		if err != nil {
			return err
		}

		err = bucket.ForEach(func(k, v []byte) error {
			buffer := bytes.NewBuffer(v)
			dec := gob.NewDecoder(buffer)
			j := new(Job)
			err := dec.Decode(j)

			if err != nil {
				return err
			}

			err = j.InitDelayDuration(false)

			if err != nil {
				return err
			}

			allJobs = append(allJobs, j)

			return nil
		})

		return err
	})

	return allJobs, err
}

func (db *BoltJobDB) Get(id string) (*Job, error) {
	j := new(Job)

	err := db.dbConn.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(jobBucket)

		v := b.Get([]byte(id))
		if v == nil {
			return fmt.Errorf("Job with id of %s not found.", id)
		}

		buffer := bytes.NewBuffer(v)
		dec := gob.NewDecoder(buffer)
		err := dec.Decode(j)

		return err
	})
	if err != nil {
		return nil, err
	}

	j.Id = id
	return j, nil
}

func (db *BoltJobDB) Delete(id string) {
	db.dbConn.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(jobBucket)
		bucket.Delete([]byte(id))
		return nil
	})
}

func (j *Job) Delete(cache JobCache, db JobDB) {
	j.Disable()
	cache.Delete(j.Id)
	db.Delete(j.Id)
}

func (db *BoltJobDB) Save(j *Job) error {
	err := db.dbConn.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(jobBucket)
		if err != nil {
			return err
		}

		buffer := new(bytes.Buffer)
		enc := gob.NewEncoder(buffer)
		err = enc.Encode(j)
		if err != nil {
			return err
		}

		err = bucket.Put([]byte(j.Id), buffer.Bytes())
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (j *Job) Save(db JobDB) error {
	return db.Save(j)
}
