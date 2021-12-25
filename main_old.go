// package main

// import (
// 	"bytes"
// 	"encoding/gob"
// 	"errors"
// 	"fmt"
// 	"net"
// 	"net/rpc"
// 	"path/filepath"
// 	"sync"
// 	"time"

// 	"github.com/boltdb/bolt"
// 	"github.com/spiral/goridge"
// )

// var (
// 	ErrNotFound = errors.New("kvstore: key not found")
// 	ErrBadValue = errors.New("kvstore: bad value")
// )

// type ColObj struct {
// 	Name string
// 	Type string
// 	Data []interface{}
// 	Pres []int
// }

// type InfoObj struct {
// 	Name string
// 	Pid string
// }

// type SetPidArgs struct {
// 	Pid string
// 	Name string
// }

// type GetPidArgs struct {
// 	Name string
// }
 
// type PutArgs struct {
// 	Key string
// 	Value ColObj
// 	Base bool
// }
// type GetArgs struct {
// 	Key string
// 	Base bool
// }

// type KVStore struct {
// 	db *bolt.DB
// 	mdb map[string]ColObj
// 	name string
// 	pid string
// 	mutex sync.Mutex
// }



// func (c *ColObj) gobEncode() ([]byte, error) {
//     buf := new(bytes.Buffer)
//     enc := gob.NewEncoder(buf)
//     err := enc.Encode(c)
//     if err != nil {
//         return nil, err
//     }
//     return buf.Bytes(), nil
// }

// func gobDecode(data []byte) (ColObj, error) {
//     var c ColObj
//     buf := bytes.NewBuffer(data)
//     dec := gob.NewDecoder(buf)
//     err := dec.Decode(&c)
//     if err != nil {
//         return ColObj{}, err
//     }
//     return c, nil
// }

// func Open(path string, name string) (*KVStore, error) {
// 	opts := &bolt.Options{
// 		Timeout: 50 * time.Millisecond,
// 	}
// 	if db, err := bolt.Open(path, 0640, opts); err != nil {
// 		return nil, err
// 	} else {
// 		// baseDB for base columns
// 		err = db.Update(func(tx *bolt.Tx) error {
// 			_, err := tx.CreateBucketIfNotExists([]byte(fmt.Sprintf("%v_base", name)))
// 			return err
// 		})
// 		if err != nil {
// 			return nil, err
// 		}
// 		mdb := map[string]ColObj{}
// 		if err != nil {
// 			return nil, err
// 		} else {
// 			return &KVStore{db: db, mdb: mdb, name: name, pid: ""}, nil
// 		}
// 	}
// }

// func (kvs *KVStore) Info(args struct{}, reply *InfoObj) error {
// 	*reply = InfoObj{kvs.name, kvs.pid}
// 	return nil
// } 

// func (kvs *KVStore) SetPID(args SetPidArgs, reply *string) error {
// 	fmt.Printf("SET PID %v (%v);", args.Pid, args.Name,)
// 	kvs.pid = args.Pid
// 	*reply = kvs.pid
// 	return nil
// } 

// func (kvs *KVStore) Put(args PutArgs, reply *bool) error {
// 	fmt.Printf("PUT %v (%v); base: %v\n", args.Key, args.Value.Name, args.Base)
// 	if !args.Base {
// 		kvs.mutex.Lock()
// 		kvs.mdb[args.Key] = args.Value
// 		kvs.mutex.Unlock()
// 		*reply = true
// 		return nil
// 	}
// 	err := kvs.db.Update(func(tx *bolt.Tx) error {
// 		kvs.mutex.Lock()
// 		b := tx.Bucket([]byte(fmt.Sprintf("%v_base", kvs.name)))
// 		enc, err := args.Value.gobEncode()
// 		if err != nil {
// 			kvs.mutex.Unlock()
// 			return fmt.Errorf("could not encode %s: %s", args.Key, err)
// 		}
// 		err = b.Put([]byte(args.Key), enc)
// 		kvs.mutex.Unlock()
// 		if err == nil {
// 			*reply = true
// 		}
// 		return err
// 	})
// 	return err
// }

// func (kvs *KVStore) Get(args GetArgs, reply *ColObj) error {
// 	fmt.Printf("GET %v; base: %v\n", args.Key, args.Base)
// 	var c ColObj
// 	if !args.Base {
// 		kvs.mutex.Lock()
// 		*reply = kvs.mdb[args.Key]
// 		kvs.mutex.Unlock()
// 		return nil
// 	}
// 	err := kvs.db.View(func(tx *bolt.Tx) error {
// 		var err error
// 		kvs.mutex.Lock()
// 		b := tx.Bucket([]byte(fmt.Sprintf("%v_base", kvs.name)))
// 		k := []byte(args.Key)
// 		c, err = gobDecode(b.Get(k))
// 		kvs.mutex.Unlock()
// 		if err != nil {
// 			return err
// 		}
// 		return nil
// 	}) 
// 	if err != nil {
// 		fmt.Printf("Could not get %s", args.Key)
// 		return err
// 	}
// 	*reply = c
// 	return nil
// }

// func (kvs *KVStore) Delete(args GetArgs, reply *bool) error {
// 	fmt.Printf("DEL %v; base: %v\n", args.Key, args.Base)
// 	if !args.Base {
// 		kvs.mutex.Lock()
// 		delete(kvs.mdb, args.Key)
// 		kvs.mutex.Unlock()
// 		*reply = true
// 		return nil
// 	}
	
// 	err := kvs.db.Update(func(tx *bolt.Tx) error {
// 		kvs.mutex.Lock()
// 		b := tx.Bucket([]byte(fmt.Sprintf("%v_base", kvs.name)))
// 		err := b.Delete([]byte(args.Key))
// 		if err == nil {
// 			*reply = true
// 		}
// 		kvs.mutex.Unlock()
// 		return err
// 	})
// 	return err
// }

// func (kvs *KVStore) List(args struct{Base bool}, reply *map[string]string) error {
// 	fmt.Println("List")
// 	results := map[string]string{}
// 	if !args.Base {
// 		kvs.mutex.Lock()
// 		for k, v := range kvs.mdb {
// 			results[k] = v.Name
// 		}
// 		kvs.mutex.Unlock()
// 		*reply = results
// 		return nil
// 	}


// 	err := kvs.db.View(func(tx *bolt.Tx) error {
// 		kvs.mutex.Lock()
// 		b := tx.Bucket([]byte(fmt.Sprintf("%v_base", kvs.name)))
// 		c := b.Cursor()

// 		for k, v := c.First(); k != nil; k, v = c.Next() {
// 			decoded, err := gobDecode(v)
// 			if err != nil {
// 					return err
// 			}
// 			results[string(k)] = decoded.Name
// 		}
// 		*reply = results
// 		kvs.mutex.Unlock()
// 		return nil
// 	})
// 	return err
// }

// func main() {
// 	ln, err := net.Listen("tcp", ":42586")
// 	if err != nil {
// 		panic(err)
// 	}
// 	abs, _ := filepath.Abs(".")
// 	fmt.Printf("server started on %v;\nin %v project\n\n", "0.0.0.0:42580", abs)

// 	store, err := Open(fmt.Sprintf("%v/.dac/data/store.db", abs), abs)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer store.db.Close()

// 	rpc.Register(store)
// 	for {
// 		conn, err := ln.Accept()
// 		if err != nil {
// 			continue
// 		}
// 		go rpc.ServeCodec(goridge.NewCodec(conn))
// 	}
// }