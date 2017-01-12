package main

import (
	"log"
	"os"
	"flag"
	"net/http"
	"encoding/json"
	exp "hls-util/explorer"
	"crypto/md5"
	"encoding/hex"
	"time"
	"hls-util/codec/hls"
	"strings"
)

var (
	port = flag.String("port", "9090", "listen port")
	sourceDir = flag.String("sourceDir", "", "source directory")
	targetDir = flag.String("targetDir", "", "target directory")
)

func main() {

	flag.Parse()
	if (*sourceDir == "" || *targetDir == "") {
		flag.PrintDefaults()
		return
	}

	log.Println("port:[", *port, "], srcDir:[", *sourceDir, "], dstDir:[", *targetDir, "]")

	hls.Init(*sourceDir, *targetDir)

	os.MkdirAll(*sourceDir, 0666);

	mux := http.NewServeMux()
	mux.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir("html"))))
	mux.HandleFunc("/playlist", handlePlayList)
	mux.HandleFunc("/list", handleList)
	go func() {
		http.ListenAndServe(":" + *port, mux)
	}()
	select {}
}

func handlePlayList(w http.ResponseWriter, r *http.Request) {
	beginTime := time.Now().Unix()
	path := r.URL.Query().Get("path")
	if path == "" {
		w.Write([]byte("'path' must not be null."))
		return
	}
	path = strings.Replace(path, "../", "/", -1)
	log.Println("play", path)

	playPath, message := hls.Transport(path)

	w.Header().Add("content-type", "application/json;charset=utf-8")

	model := make(map[string]interface{})
	model["result"] = playPath
	model["message"] = message
	if (message == "already exists" || message == "transcoding") && playPath != "" {
		model["statusCode"] = 200
	} else {
		model["statusCode"] = 201
	}
	model["elapsedTime"] = time.Now().Unix() - beginTime

	b, err := json.Marshal(model)
	if err != nil {
		log.Println(err)
		w.Write([]byte(`{"statusCode":201,"message":"` + err.Error() + `"}`))
		return
	}
	w.Write(b)
}

func handleList(w http.ResponseWriter, r *http.Request) {
	beginTime := time.Now().Unix()
	path := r.URL.Query().Get("path")
	path = strings.Replace(path, "../", "/", -1)
	log.Println("list path", path)
	list, err := exp.ListDir(*sourceDir + string(os.PathSeparator) + path)

	for i, item := range list {

		re := item.Path[len(*sourceDir):]
		md := md5.New().Sum([]byte(re))
		key := hex.EncodeToString(md);

		item.Id = key
		item.Path = re

		list[i] = item
	}

	w.Header().Add("content-type", "application/json;charset=utf-8")

	model := make(map[string]interface{})
	model["result"] = list
	if err != nil {
		model["statusCode"] = 201
		model["message"] = err.Error()
	} else {
		model["statusCode"] = 200
		model["message"] = "success"
	}
	model["elapsedTime"] = time.Now().Unix() - beginTime

	b, err := json.Marshal(model)
	if err != nil {
		log.Println(err)
		w.Write([]byte(`{"statusCode":201,"message":"` + err.Error() + `"}`))
		return
	}
	w.Write(b)
}
