package explorer

import (
	"log"
	"io/ioutil"
	"os"
	"strings"
	"errors"
)

const sep = string(os.PathSeparator)

func ListDir(path string) ([]MediaItem, error) {
	list := []MediaItem{}

	f, err := os.Stat(path)
	if err != nil || !f.IsDir() {
		return nil, errors.New("The file not a directory")
	}

	info, err := ioutil.ReadDir(path)
	if err != nil {
		log.Println(err)
		return nil, errors.New("Can't read the directory '" + path + "'")
	}

	for _, f := range info {
		if strings.HasPrefix(f.Name(), ".") {
			continue
		}
		item := MediaItem{}

		item.Path = path + sep + f.Name()
		for strings.Index(item.Path, (sep + sep)) >= 0 {
			item.Path = strings.Replace(item.Path, (sep + sep), sep, -1)
		}
		item.Name = f.Name()
		item.Mode = f.Mode().String()
		item.Size = f.Size()

		index := strings.LastIndex(item.Name, ".");
		if index > 0 && index < len(item.Name) && !f.IsDir() {
			item.Suffix = item.Name[index + 1:]
		}

		item.UpdateTime = f.ModTime().Unix() * 1000

		list = append(list, item)
	}
	return list, nil
}
