package zkClient

import (
	"fmt"
	"log"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"stathat.com/c/consistent"
)

type Client struct {
	zookeeperConn *zk.Conn
	hash          *consistent.Consistent
}

func (z *Client) Connect(zkIP string) {
	var err error
	z.zookeeperConn, _, err = zk.Connect([]string{zkIP}, time.Second) //Default zkIP = "127.0.0.1:2181"
	handleError(err)
	fmt.Println("[zookeeper] : Connected with Zookeeper...")
}

func New() *Client {

	return &Client{}
}

func (z *Client) GetHash() *consistent.Consistent {
	return z.hash
}

func (z *Client) GetWatch(path string) {
	snapshots, errors := z.setWatch(path)
	fmt.Println("[zookeeper] : Set watch on path " + path)
	go func() {
		for {
			select {
			case snapshot := <-snapshots:
				z.hash = consistent.New()
				for i := 0; i < len(snapshot); i++ {
					res, _, err := z.zookeeperConn.Get(path + "/" + snapshot[i])
					if err != nil {
						log.Println("error : ", err)
						continue
					}
					z.hash.Add(string(res[:]))
				}
			case err := <-errors:
				panic(err)
			}
		}
	}()

}

func (z *Client) setWatch(path string) (chan []string, chan error) {
	snapshots := make(chan []string)
	errors := make(chan error)
	go func() {
		for {
			snapshot, _, events, err := z.zookeeperConn.ChildrenW(path)
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
