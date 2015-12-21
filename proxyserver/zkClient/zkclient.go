package zkClient

import (
	"fmt"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"stathat.com/c/consistent"
)

var zookeeperConn *zk.Conn
var HTTPhash *consistent.Consistent
var PushpinHash *consistent.Consistent
var err error

func Connect() {
	zookeeperConn, _, err = zk.Connect([]string{"127.0.0.1:2181"}, time.Second)
	handleError(err)
	time.Sleep(time.Second)
	fmt.Println("[zookeeper] : Connected with Zookeeper...")
}

func GetHash() (*consistent.Consistent, *consistent.Consistent) {
	HTTPhash = consistent.New()
	PushpinHash = consistent.New()
	return HTTPhash, PushpinHash
}

func GetWatch(path string, hash *consistent.Consistent) {
	snapshots, errors := setWatch(zookeeperConn, path)
	fmt.Println("[zookeeper] : Set watch on path " + path)
	go func() {
		for {
			select {
			case snapshot := <-snapshots:
				//hash = consistent.New()
				for i := 0; i < len(snapshot); i++ {
					res, _, _ := zookeeperConn.Get(path + "/" + snapshot[i])
					hash.Add(string(res[:]))
					fmt.Println(PushpinHash.Members())
				}
			case err := <-errors:
				panic(err)
			}
		}
	}()

}

func setWatch(conn *zk.Conn, path string) (chan []string, chan error) {
	snapshots := make(chan []string)
	errors := make(chan error)
	go func() {
		for {
			snapshot, _, events, err := conn.ChildrenW(path)
			if err != nil {
				errors <- err
				return
			}
			snapshots <- snapshot
			evt := <-events
			if evt.Err != nil {
				errors <- evt.Err
				return
			}
		}
	}()
	return snapshots, errors
}

//handleError stops the program execution
func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
