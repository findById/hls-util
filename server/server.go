package main

import (
	"os/exec"
	"log"
	"os"
	"flag"
	"net/http"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"bufio"
	"strings"
	"encoding/json"
)

const hls string = "ffmpeg -i %s -c:v libx264 -codec:a mp3 -hls_time 10 -hls_list_size 0 %s";

var (
	sourceDir = flag.String("sourceDir", "", "source directory")
	targetDir = flag.String("targetDir", "", "target directory")
)

func main() {
	execute("uname -a")
	execute("pwd")
	flag.Parse()
	if (*sourceDir == "" || *targetDir == "") {
		flag.PrintDefaults()
		return
	}

	os.MkdirAll(*sourceDir, 0666);

	mux := http.NewServeMux()
	mux.HandleFunc("/decode", handler)
	mux.HandleFunc("/list", handleList)
	go func() {
		http.ListenAndServe(":9090", mux)
	}()
	select {}
}

func execute(command string) int {
	log.Println(command)
	cmd := exec.Command("/bin/sh", "-c", command)
	buff, err := cmd.Output();
	if err != nil {
		log.Println(err)
		return -1
	}
	log.Println(string(buff))
	return 1
}

func handleList(w http.ResponseWriter, r *http.Request) {

	f, _ := os.Open("list")
	defer f.Close()
	buffer, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
		w.Write([]byte("Can't reading list"))
		return
	}
	br := bufio.NewReader(strings.NewReader(string(buffer)))
	array := make([]MediaItem, 0)
	for {
		temp, _, err := br.ReadLine()
		if err != nil {
			break
		}
		line := strings.TrimSpace(string(temp))
		// Empty line
		if line == "" {
			continue
		}

		index := strings.Index(line, ":")
		if index > 0 && index < len(line) {
			key := strings.TrimSpace(line[0:index])
			value := strings.TrimSpace(line[index + 1:])
			log.Println(key, value)
			item := MediaItem{}
			err := json.Unmarshal([]byte(value), &item)
			if err != nil {
				log.Println(err)
				continue
			}

			array = append(array, item)
		}
	}

	w.Header().Add("content-type", "text/html;charset=utf-8")

	b, err := json.Marshal(array)
	if err != nil {
		log.Println(err)
		w.Write([]byte("Can't reading list"))
		return
	}
	w.Write(b)
}

func handler(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("source")
	if source == "" {
		w.Write([]byte("source must not be null."))
		return
	}
	md := md5.New().Sum([]byte(source))
	key := hex.EncodeToString(md);

	os.MkdirAll(*targetDir + "/" + key, 0777)

	if isExists(*targetDir + "/" + key + "/success") {
		log.Println("already exists")
		w.Write([]byte(key))
		// return;
	}
	index := strings.LastIndex(source, "/")
	dot := strings.LastIndex(source, ".")
	filename := ""
	if index > 0 && index < len(source) {
		filename = source[index:dot]
	} else {
		filename = source[:dot]
	}
	hlsPath := "/" + key + "/" + filename + ".m3u8"

	status := execute(fmt.Sprintf(hls, *sourceDir + "/" + source, *targetDir + hlsPath))
	if status == 1 {
		ioutil.WriteFile(*targetDir + "/" + key + "/success", []byte("success"), 0666)

		item := new(MediaItem)
		item.Id = key
		item.SourcePath = source
		item.HlsPath = hlsPath
		item.Filename = filename

		appendList(source, item)
	}

	w.Header().Add("content-type", "text/html;charset=utf-8")

	model := make(map[string]string)
	model[source] = key + "/playlist.m3u8"
	b, err := json.Marshal(model)
	if err != nil {
		log.Println(err)
		w.Write([]byte("Can't reading list"))
		return
	}
	w.Write(b)
}

type MediaItem struct {
	Id         string        `json:"id"`
	Filename   string        `json:"filename"`
	SourcePath string        `json:"sourcePath"`
	HlsPath    string        `json:"hlsPath"`
}

func appendList(key string, item *MediaItem) {
	b, err := json.Marshal(item)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(key, string(b))
	f, err := os.OpenFile("list", os.O_CREATE | os.O_APPEND | os.O_WRONLY, 0666)
	if err != nil {
		log.Println(err)
		return
	}
	defer f.Close()

	f.WriteString(item.Id + " : " + string(b) + "\r\n")
}

func isExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}