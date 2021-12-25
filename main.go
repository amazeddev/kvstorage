package main

import (
	"errors"
	"fmt"
	"net"
	"net/rpc"
	"path/filepath"
	"sync"

	"github.com/spiral/goridge"
)

var (
	ErrNotFound = errors.New("kvstore: key not found")
	ErrBadValue = errors.New("kvstore: bad value")
)

// type ColObj struct {
// 	Name string
// 	Type string
// 	Data []interface{}
// 	Pres []int
// }

type InfoObj struct {
	Name string
	Pid string
}

type SetPidArgs struct {
	Pid string
	Name string
}

type GetPidArgs struct {
	Name string
}
 
type PutArgs struct {
	Key string
	Value interface{}
	Base bool
}
type GetArgs struct {
	Key string
	Base bool
}

type KVStore struct {
	mdb map[string]interface{}
	name string
	pid string
	mutex sync.Mutex
}

func Open(name string) (*KVStore, error) {
	mdb := map[string]interface{}{}
	
	return &KVStore{mdb: mdb, name: name, pid: ""}, nil
}

func (kvs *KVStore) Info(args struct{}, reply *InfoObj) error {
	*reply = InfoObj{kvs.name, kvs.pid}
	return nil
} 

func (kvs *KVStore) SetPID(args SetPidArgs, reply *string) error {
	fmt.Printf("SET PID %v (%v);", args.Pid, args.Name,)
	kvs.pid = args.Pid
	*reply = kvs.pid
	return nil
} 

func (kvs *KVStore) Put(args PutArgs, reply *bool) error {
	fmt.Printf("PUT %v;\n", args.Key)
	
	kvs.mutex.Lock()
	kvs.mdb[args.Key] = args.Value
	kvs.mutex.Unlock()
	*reply = true
	return nil
}

func (kvs *KVStore) Get(args GetArgs, reply *interface{}) error {
	fmt.Printf("GET %v;\n", args.Key)
	
	kvs.mutex.Lock()
	*reply = kvs.mdb[args.Key]
	kvs.mutex.Unlock()
	return nil
}

func (kvs *KVStore) Delete(args GetArgs, reply *bool) error {
	fmt.Printf("DEL %v;\n", args.Key)
	
	kvs.mutex.Lock()
	delete(kvs.mdb, args.Key)
	kvs.mutex.Unlock()
	*reply = true
	return nil
}

func (kvs *KVStore) List(args struct{}, reply *[]string) error {
	fmt.Println("List")
	results := []string{}

	kvs.mutex.Lock()
	for k, _ := range kvs.mdb {
		results = append(results, k)
	}
	kvs.mutex.Unlock()
	*reply = results
	return nil
}

func main() {
	ln, err := net.Listen("tcp", ":42586")
	if err != nil {
		panic(err)
	}
	abs, _ := filepath.Abs(".")
	fmt.Printf("server started on %v;\nin %v project\n\n", "0.0.0.0:42580", abs)

	store, err := Open(abs)
	if err != nil {
		panic(err)
	}

	rpc.Register(store)
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go rpc.ServeCodec(goridge.NewCodec(conn))
	}
}