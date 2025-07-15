/*
 * @Author: FunctionSir
 * @License: AGPLv3
 * @Date: 2025-04-15 20:01:47
 * @LastEditTime: 2025-07-15 18:48:51
 * @LastEditors: FunctionSir
 * @Description: -
 * @FilePath: /any-ecs-doh-proxy/main.go
 */

package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/FunctionSir/goset"
	"github.com/FunctionSir/readini"
)

type SrvStatus struct {
	StartedAt  int64
	TotQueries atomic.Uint64
}

var Status SrvStatus

// Global config, do not modify it after init.
var Config readini.Conf
var DusNeeded goset.Set[string]

// Load config.
func loadConf() readini.Conf {
	if len(os.Args) <= 1 {
		panic("you need to specify the config file you want to use")
	}
	conf, err := readini.LoadFromFile(os.Args[1])
	if err != nil {
		panic(err)
	}
	return conf
}

// Load dynamic upstream needed sites.
func loadDusNeeded() {
	DusNeeded = make(goset.Set[string])
	if !Config.HasKey("options", "DusNeeded") {
		return
	}
	buf, err := os.ReadFile(Config["options"]["DusNeeded"])
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(buf), "\n")
	for _, tmp := range lines {
		site := strings.TrimSpace(tmp)
		if len(site) <= 0 {
			continue
		}
		DusNeeded.Insert(site)
	}
}

func chkConf() {
	if !Config.HasKey("options", "IpDb") {
		panic("key IpDb in section options not found")
	}
	if !Config.HasKey("options", "Listen") {
		panic("key Listen in section options not found")
	}
	if !Config.HasKey("options", "Cert") || !Config.HasKey("options", "Key") {
		fmt.Println("Warning: HTTPS will be disabled due to incomplete config, that might be insecure!")
	}
	if !Config.HasKey("options", "Upstream") {
		panic("no upstream specified")
	}
	if Config.HasKey("options", "Proxy") {
		os.Setenv("HTTP_PROXY", Config["options"]["Proxy"])
		os.Setenv("HTTPS_PROXY", Config["options"]["Proxy"])
		fmt.Println("Info: using proxy " + Config["options"]["Proxy"] + ".")
	}
}

func main() {
	fmt.Println("DoH Proxy With Customizable ECS Support Server")
	fmt.Println("By FunctionSir | Feel free to use under AGPLv3")
	Status.TotQueries.And(0)
	Config = loadConf()
	chkConf()
	loadDusNeeded()
	DbOpen()
	DbPrepare()
	fmt.Println("Info: will listen " + Config["options"]["Listen"] + "...")
	Status.StartedAt = time.Now().Unix()
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/{CountryCode}/{Province}/{City}", queryHandler)
	var err error
	if Config.HasKey("options", "Cert") && Config.HasKey("options", "Key") {
		err = http.ListenAndServeTLS(Config["options"]["Listen"], Config["options"]["Cert"], Config["options"]["Key"], nil)
	} else {
		err = http.ListenAndServe(Config["options"]["Listen"], nil)
	}
	panic(err)
}
