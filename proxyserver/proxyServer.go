package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/sacheendra/api_frontend/config"
	"github.com/sacheendra/api_frontend/infrastructure"
)

var PushpinServerAddr string
var HTTPServerAddr string
var client *http.Client

func main() {
	serverInstance := Init()
	fmt.Println("running server...")
	fmt.Println("listening on " + serverInstance.Addr)
	log.Fatal(serverInstance.ListenAndServe())
}

func Init() *http.Server {

	currentConfig := getCurrentConfig()
	PushpinServerAddr = currentConfig.PushpinServers.ProxyAddress
	HTTPServerAddr = strconv.Itoa(currentConfig.HTTPServer.Port)
	client = &http.Client{}
	fmt.Println("Initializing server..." + PushpinServerAddr)
	fmt.Println("Environment: " + currentConfig.Name)
	return &http.Server{
		Addr:           ":7000",
		Handler:        http.HandlerFunc(routeReq),
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
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
		path = "http://" + user + ":" + pass + "@localhost:" + HTTPServerAddr + req.URL.String()
		fmt.Println("Started " + req.Method + " " + req.URL.String())
	}

	remoteReq, err := http.NewRequest(req.Method, path, req.Body)

	if err != nil {
		log.Println("error : ", err)
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
		panic(err)
	}
}
