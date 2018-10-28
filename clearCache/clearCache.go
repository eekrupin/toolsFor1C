package main

import (
	"bufio"
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strings"
	"unicode/utf8"
)

const (
	BOM        = rune(65279)
	PATHIBASES = "\\AppData\\Roaming\\1C\\1CEStart\\ibases.v8i"
	PATHCACHE  = "\\AppData\\Local\\1C\\1cv8\\"
)

var userHomeDir string

func init() {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	userHomeDir = usr.HomeDir
}

type baseListItem struct {
	name    string
	id      string
	connect string
}

//Удаляет кэш 1С по идентификатору в списке баз. Параметры передавать в кавычках, пример:clearCache.exe "WRF_prod" "GOT_dev"
func main() {

	chosenBases := make(map[string]interface{})
	for _, baseName := range os.Args[1:] {
		chosenBases[strings.ToLower(baseName)] = struct{}{}
	}

	baseList := getBaseListOneS(chosenBases)

	for _, item := range baseList {
		err := os.RemoveAll(userHomeDir + PATHCACHE + item.id)
		if err != nil {
			log.Fatalf("Delete folder fails: name = %s, id = %s. Error: %s", item.name, item.id, err)
		}
	}
}
func getBaseListOneS(chosenBases map[string]interface{}) []*baseListItem {

	list := make([]*baseListItem, 0)

	fileData, err := ioutil.ReadFile(userHomeDir + PATHIBASES)
	if err != nil {
		log.Fatalf("Error reading base list: %s", err)
	}

	scanner := bufio.NewScanner(bytes.NewReader(fileData))
	var currentItem *baseListItem
	searchNext := true
	for scanner.Scan() {
		line := scanner.Text()
		lineLen := utf8.RuneCountInString(line)
		runes := []rune(line)
		if runes[0] == BOM {
			runes = runes[1:]
			lineLen--
		}

		if lineLen > 0 && string(runes[:1]) == "[" {
			name := strings.ToLower(string(runes[1 : lineLen-1]))
			if chosenBases[name] != nil {
				searchNext = false
				currentItem = &baseListItem{}
				currentItem.name = name
				list = append(list, currentItem)
			} else {
				searchNext = true
			}
		}

		if searchNext {
			continue
		}

		indexSeparator := strings.Index(line, "=")
		if indexSeparator == -1 {
			continue
		}
		key := line[:indexSeparator]
		value := line[indexSeparator+1:]
		switch key {
		case "ID":
			currentItem.id = value
		case "Connect":
			currentItem.connect = value
		}
	}
	return list
}
