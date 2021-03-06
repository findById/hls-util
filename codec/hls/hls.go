package hls

import (
	"log"
	"os/exec"
	"crypto/md5"
	"encoding/hex"
	"os"
	"strings"
	"time"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

const (
	hls string = "ffmpeg -i %s -c:v libx264 -c:a mp3 -hls_time 10 -hls_list_size 0 %s";
)

var (
	VIDEO_SUFFIX = []string{".avi", ".mkv", ".mp4", ".rmvb", ".rm", ".flv", ".mov", ".vob", ".wmv", ".ts"}
	srcRootDir string = ""
	dstRootDir string = ""
)

func Init(srcDir, dstDir string) {
	srcRootDir = srcDir
	dstRootDir = dstDir
}

func IsVideo(ext string) bool {
	for _, su := range VIDEO_SUFFIX {
		if ext == su {
			return true
		}
	}
	return false
}

func Transport(path string) (string, string) {
	if path == "" {
		return "", "'path' must not be null"
	}
	f, err := os.Stat(srcRootDir + "/" + path)
	if err != nil || f.IsDir() {
		return "", "Can't handle folder"
	}
	suffix := strings.ToLower(filepath.Ext(path))
	if !IsVideo(suffix) {
		return "", "Not a video file"
	}

	md := md5.New().Sum([]byte(path))
	key := hex.EncodeToString(md);

	os.MkdirAll(dstRootDir + "/" + key, 0777)

	index := strings.LastIndex(path, "/")
	dot := strings.LastIndex(path, ".")
	filename := ""
	if index > 0 && index < len(path) {
		filename = path[index:dot]
	} else {
		filename = path[:dot]
	}
	hlsPath := "/" + key + "/" + filename + ".m3u8"

	if isExists(dstRootDir + "/" + key + "/success") {
		log.Println("already exists")
		return hlsPath, "already exists"
	}

	lockFile := dstRootDir + "/" + key + "/success"
	ioutil.WriteFile(lockFile, []byte("success"), 0666)
	go func() {
		beginTime := time.Now().Unix()
		status := execute(fmt.Sprintf(hls, srcRootDir + "/" + path, dstRootDir + hlsPath))
		if !status {
			execute("rm -rf " + lockFile)
			return
		}
		log.Println("transcode success")
		total := time.Now().Unix() - beginTime;
		log.Println("used", total, "s")
	}()

	return hlsPath, "transcoding"
}

func execute(command string) bool {
	log.Println(command)
	cmd := exec.Command("/bin/sh", "-c", command)
	_, err := cmd.Output();
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func isExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}