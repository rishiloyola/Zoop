package pipe

import (
	"fmt"
	"net/url"

	"stathat.com/c/consistent"
	"zkClient"
)

type PipeClient struct{}

var zkStreamClient *zkClient.Client
var zkHttpClient *zkClient.Client

func (p *PipeClient) Init(zkHTTPpath string, zkStreampath string, zkIP string) {
	zkStreamClient = zkClient.New()
	zkHttpClient = zkClient.New()
	zkStreamClient.Connect(zkIP)
	zkHttpClient.Connect(zkIP)
	zkHttpClient.GetWatch(zkHTTPpath)
	zkStreamClient.GetWatch(zkStreampath)
}

func (p *PipeClient) GetServers(path *url.URL) *consistent.Consistent {
	streamReq := identifyReqType(path)
	if streamReq {
		return zkStreamClient.GetHash()
	} else {
		return zkHttpClient.GetHash()
	}

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
