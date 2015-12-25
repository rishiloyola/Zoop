package proxyserver

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"pipe"
	"stathat.com/c/consistent"
)

type proxyServer struct {
	client       *http.Client
	httpServer   *http.Server
	pclient      *pipe.PipeClient
	zkclient     *zk.Conn
	zkHttpPath   string
	zkStreamPath string
}

func (s *proxyServer) Init(zkHttpPath string, zkStreamPath string, zkIP string) {
	var err error

	s.client = &http.Client{}
	s.zkclient, _, err = zk.Connect([]string{zkIP}, time.Second)
	handleError(err)
	time.Sleep(time.Second)
	s.httpServer = &http.Server{
		Addr:           ":7000",
		Handler:        http.HandlerFunc(s.routeReq),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.pclient.Init(zkHttpPath, zkStreamPath, zkIP)
	s.zkHttpPath = zkHttpPath
	s.zkStreamPath = zkStreamPath
}

func (s *proxyServer) Run() {
	fmt.Println("listening on " + s.httpServer.Addr)
	log.Fatal(s.httpServer.ListenAndServe())
}

func New() *proxyServer {
	return &proxyServer{}
}

func (s *proxyServer) routeReq(w http.ResponseWriter, req *http.Request) {

	servers := s.pclient.GetServers(req.URL)
	user, pass, _ := req.BasicAuth()
	serverAddr, err := s.getRedirectServer(servers, user)

	if err != nil {
		log.Println(err)
	}

	path := "http://" + user + ":" + pass + "@" + serverAddr + req.URL.String()
	fmt.Println("started " + req.Method + " " + req.URL.String())
	remoteReq, err := http.NewRequest(req.Method, path, req.Body)

	if err != nil {
		fmt.Println(err)
	}

	remoteResp, err := s.client.Do(remoteReq)

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

func (s *proxyServer) getRedirectServer(servers *consistent.Consistent, key string) (string, error) {

	serverAddr, err := servers.Get(key)
	return serverAddr, err

}

//handleError stops the program execution
func handleError(err error) {
	if err != nil {
		panic(err)
	}
}
