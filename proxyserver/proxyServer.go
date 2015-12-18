package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/sacheendra/api_frontend/config"
	"github.com/sacheendra/api_frontend/infrastructure"
	"github.com/samuel/go-zookeeper/zk"
	"stathat.com/c/consistent"
)

var PushpinServerAddr string

//var HTTPServerAddr string
var client *http.Client
var zookeeperConn *zk.Conn
var hash *consistent.Consistent

func main() {
	serverInstance := Init()
	fmt.Println("listening on " + serverInstance.Addr)
	log.Fatal(serverInstance.ListenAndServe())
}

func Init() *http.Server {

	currentConfig := getCurrentConfig()
	PushpinServerAddr = currentConfig.PushpinServers.ProxyAddress
	//HTTPServerAddr = strconv.Itoa(currentConfig.HTTPServer.Port)
	client = &http.Client{}
	fmt.Println("Environment: " + currentConfig.Name)
	zkconnectClient()
	getWatch()
	return &http.Server{
		Addr:           ":7000",
		Handler:        http.HandlerFunc(routeReq),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func getWatch() {
	snapshots, errors := setWatch(zookeeperConn, "/HTTPserver")
	fmt.Println("[zookeeper] : Set watch on path /HTTPserver")
	go func() {
		for {
			select {
			case snapshot := <-snapshots:
				hash = consistent.New()
				for i := 0; i < len(snapshot); i++ {
					res, _, _ := zookeeperConn.Get("/HTTPserver/" + snapshot[i])
					hash.Add(string(res[:]))
				}
			case err := <-errors:
				panic(err)
			}
		}
	}()

}

func zkconnectClient() {
	var err error
	zookeeperConn, _, err = zk.Connect([]string{"127.0.0.1:2181"}, time.Second)
	handleError(err)
	time.Sleep(time.Second)
	// path, err := zookeeperConn.Create("/localserver", []byte("data"), int32(zk.FlagEphemeral), zk.WorldACL(zk.PermAll))
	// handleError(err)
	// time.Sleep(time.Second)
	fmt.Println("[zookeeper] : Connected with Zookeeper...")
}

func getCurrentConfig() config.EnvironmentConfig {

	configRepository := infrastructure.NewConfigFileRepository(
		strings.Join([]string{os.Getenv("GOPATH"), "/src/github.com/sacheendra/api_frontend/api_frontend.json"}, ""),
	)

	systemConfig, err := configRepository.GetSystemConfiguration()
	handleError(err)

	// ---- Pause app for network to initialize during production ----
	if systemConfig.CurrentEnvironment == "PRODUCTION" {
		time.Sleep(time.Second * 1)
	}

	currentConfig, err := systemConfig.GetCurrentEnvironmentConfig()
	handleError(err)

	return currentConfig

}

func routeReq(w http.ResponseWriter, req *http.Request) {

	streamReq := identifyReqType(req.URL)
	user, pass, _ := req.BasicAuth()
	var path string

	if streamReq {
		path = "http://" + user + ":" + pass + "@" + PushpinServerAddr + req.URL.String()
		fmt.Println("started " + req.Method + " " + req.URL.String())
	} else {
		server, err := hash.Get(user)
		if err != nil {
			log.Fatal(err)
		}
		path = "http://" + user + ":" + pass + "@" + server + req.URL.String()
		fmt.Println("Started " + req.Method + " " + req.URL.String())
	}

	remoteReq, err := http.NewRequest(req.Method, path, req.Body)

	if err != nil {
		fmt.Println(err)
	}

	remoteResp, err := client.Do(remoteReq)

	if err != nil {
		log.Println("error : ", err)
	}

	fmt.Println("Redirected on " + path)
	defer remoteResp.Body.Close()
	content, err := ioutil.ReadAll(remoteResp.Body)

	if err != nil {
		log.Println("error : ", err)
	}

	w.Write(content)
	fmt.Println("Completed " + remoteResp.Status)
}

func identifyReqType(path *url.URL) bool {
	query := path.Query()
	if len(query) == 0 {
		return false
	} else {
		if _, ok := query["stream"]; ok {
			return true
		} else {
			return false
		}
	}
}

//handleError stops the program execution
func handleError(err error) {
	if err != nil {
		zookeeperConn.Close()
		panic(err)
	}
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
