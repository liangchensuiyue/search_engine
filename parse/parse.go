package parse

import (
	"bufio"
	"os"
	"strings"

	"github.com/wangbin/jiebago"
)

var seg jiebago.Segmenter
var stopWords []string

func init() {
	seg.LoadDictionary("dictionary.txt")
	loadStopWords()
}
func loadStopWords() {
	file, err := os.Open("d:\\gaodongsheng\\goproject\\src\\go_code\\search_engine\\parse\\stop_words.utf8")
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(file)
	for {
		data, _, err := reader.ReadLine()
		if err != nil {
			break
		}
		stopWords = append(stopWords, strings.TrimSpace(string(data)))
	}
}

func IsStopWords(str string) bool {
	for _, v := range stopWords {
		if str == v {
			return true
		}
	}
	return false
}
func ParseString(target string) map[string]int {
	var maps map[string]int = make(map[string]int, 0)
	chanel := seg.CutForSearch(target, true)
	for v := range chanel {
		if !IsStopWords(v) {
			_, ok := maps[v]
			if !ok {
				maps[v] = 1
			} else {
				maps[v] = maps[v] + 1
			}
		}
	}
	return maps
}
