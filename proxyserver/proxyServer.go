package proxyserver

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"

	"stathat.com/c/consistent"
	"zoop/proxyserver/zkClient"
)

type proxyServer struct {
	client      *http.Client
	HTTPhash    *consistent.Consistent
	PushpinHash *consistent.Consistent
	httpServer  *http.Server
}

func (s *proxyServer) Init() {

	s.client = &http.Client{}
	s.HTTPhash, s.PushpinHash = zkClient.GetHash()
	zkClient.Connect()
	zkClient.GetWatch("/HTTPserver", s.HTTPhash)
	zkClient.GetWatch("/pushpin", s.PushpinHash)
	s.httpServer = &http.Server{
		Addr:           ":7000",
		Handler:        http.HandlerFunc(s.routeReq),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func (s *proxyServer) Run() {
	fmt.Println("listening on " + s.httpServer.Addr)
	log.Fatal(s.httpServer.ListenAndServe())
}

func New() *proxyServer {
	return &proxyServer{}
}

func (s *proxyServer) routeReq(w http.ResponseWriter, req *http.Request) {

	streamReq := identifyReqType(req.URL)
	user, pass, _ := req.BasicAuth()
	var path string
	if streamReq {
		PushpinServerAddr, err := s.PushpinHash.Get(user)
		if err != nil {
			log.Println(err)
		}
		path = "http://" + user + ":" + pass + "@" + PushpinServerAddr + req.URL.String()
		fmt.Println("started " + req.Method + " " + req.URL.String())
	} else {
		HTTPserverAddr, err := s.HTTPhash.Get(user)
		if err != nil {
			log.Println(err)
		}
		path = "http://" + user + ":" + pass + "@" + HTTPserverAddr + req.URL.String()
		fmt.Println("Started " + req.Method + " " + req.URL.String())
	}

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
