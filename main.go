package main

import (
	"zoop/proxyserver"
)

// type proxyServer struct {
// 	client        *http.Client
// 	zookeeperConn *zk.Conn
// 	HTTPhash      *consistent.Consistent
// 	PushpinHash   *consistent.Consistent
// }

func main() {
	appserver := proxyserver.New()
	appserver.Init()
	appserver.Run()
}

// func (s *proxyServer) Init() *http.Server {

// 	client = &http.Client{}
// 	HTTPhash = consistent.New()
// 	PushpinHash = consistent.New()
// 	zkconnectClient()
// 	getWatch("/HTTPserver", HTTPhash)
// 	getWatch("/pushpin", PushpinHash)
// 	return &http.Server{
// 		Addr:           ":7000",
// 		Handler:        http.HandlerFunc(routeReq),
// 		ReadTimeout:    10 * time.Second,
// 		WriteTimeout:   10 * time.Second,
// 		MaxHeaderBytes: 1 << 20,
// 	}
// }

// func getWatch(path string, hash *consistent.Consistent) {
// 	snapshots, errors := setWatch(zookeeperConn, path)
// 	fmt.Println("[zookeeper] : Set watch on path " + path)
// 	go func() {
// 		for {
// 			select {
// 			case snapshot := <-snapshots:
// 				//hash = consistent.New()
// 				for i := 0; i < len(snapshot); i++ {
// 					res, _, _ := zookeeperConn.Get(path + "/" + snapshot[i])
// 					hash.Add(string(res[:]))
// 					fmt.Println(PushpinHash.Members())
// 				}
// 			case err := <-errors:
// 				panic(err)
// 			}
// 		}
// 	}()

// }

// func zkconnectClient() {
// 	var err error
// 	zookeeperConn, _, err = zk.Connect([]string{"127.0.0.1:2181"}, time.Second)
// 	handleError(err)
// 	time.Sleep(time.Second)
// 	fmt.Println("[zookeeper] : Connected with Zookeeper...")
// }

// func setWatch(conn *zk.Conn, path string) (chan []string, chan error) {
// 	snapshots := make(chan []string)
// 	errors := make(chan error)
// 	go func() {
// 		for {
// 			snapshot, _, events, err := conn.ChildrenW(path)
// 			if err != nil {
// 				errors <- err
// 				return
// 			}
// 			snapshots <- snapshot
// 			evt := <-events
// 			if evt.Err != nil {
// 				errors <- evt.Err
// 				return
// 			}
// 		}
// 	}()
// 	return snapshots, errors
// }
