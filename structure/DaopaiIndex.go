package structure

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"go_code/search_engine/parse"
	"go_code/search_engine/utils"
)

type DaoPaiElment struct {
	DocumentId int // <...,...,...>
	TF         int
	IsTitle    int8 // 标题中是否出现 title
}
type DaoPaiIndex struct {
	Dictionary *WordTree
	DaoPaiFile *os.File
	Lock       *sync.Mutex
	IsDisk     bool
	Count      int
}

func (daoPaiIndex *DaoPaiIndex) Clean() {
	daoPaiIndex.Dictionary.Root = nil
	daoPaiIndex.DaoPaiFile = nil
}
func (daoPaiIndex *DaoPaiIndex) GetList(word string) ([]*DaoPaiElment, error) {
	if daoPaiIndex.IsDisk {
		node, err := daoPaiIndex.Dictionary.Search(word)
		if err != nil {
			return []*DaoPaiElment{}, err
		}
		return daoPaiIndex.GetDaoPaiListByIndex(node.Index)
	} else {
		return daoPaiIndex.Dictionary.GetList(word)
	}
}
func (daoPaiIndex *DaoPaiIndex) LoadDaoPaiFile(origin string) {
	daoPaiIndex.Clean()
	var err error
	var index int = 0
	daoPaiIndex.DaoPaiFile, err = os.Open(origin)
	// reader := bufio.NewReader(daoPaiIndex.DaoPaiFile)

	cache := make([]byte, 1024)
	wordByte := make([]byte, 0)
	if err != nil {
		panic(err)
	}
	for {
		n, err := daoPaiIndex.DaoPaiFile.Read(cache)
		if err != nil {
			break
		}
		for i := 0; i < n; i++ {
			if cache[i] == ':' {
				word := string(wordByte)
				daoPaiIndex.Dictionary.InsertIndexNode(word, index)
				daoPaiIndex.Count++
				size := utils.BytesToInt(cache[i+1 : i+9])
				fmt.Println(index, word, size)
				daoPaiIndex.DaoPaiFile.Seek(int64(index+len(wordByte)+1+8)+size, 0)
				index = int(int64(index+len(wordByte)+1+8) + size)
				wordByte = make([]byte, 0)
				break
			} else {
				wordByte = append(wordByte, cache[i])
			}
		}
		// data, _, err := reader.ReadLine()
		// if err != nil {
		// 	break
		// }
		// for i := 0; i < len(data); i++ {
		// 	if data[i] == ':' {
		// 		word := string(data[:i])
		// 		daoPaiIndex.Dictionary.InsertIndexNode(word, index)
		// 		break
		// 	}
		// }
		// index += len(data)
	}
}
func (daoPaiIndex *DaoPaiIndex) ParseArticle(id int, title, content string) bool {
	maps := make(map[string]*DaoPaiElment, 0)
	wordmap1 := parse.ParseString(title)
	for p, _ := range wordmap1 {
		maps[p] = &DaoPaiElment{
			DocumentId: id,
			TF:         1,
			IsTitle:    1,
		}
	}

	wordmap2 := parse.ParseString(content)
	for p, v := range wordmap2 {
		_, ok := maps[p]
		if !ok {
			maps[p] = &DaoPaiElment{
				DocumentId: id,
				TF:         v,
				IsTitle:    0,
			}
		} else {
			a := maps[p]
			a.TF = v
			// maps[p].TF = v
		}

	}
	bools := false
	for p, v := range maps {
		fmt.Println(p)
		daoPaiIndex.Count++
		daoPaiIndex.Dictionary.InsertListNode(p, v)
		if daoPaiIndex.Count > 20 {
			bools = true
		}
	}
	return bools
}
func (daoPaiIndex *DaoPaiIndex) GetWordByIndex(index int) (string, error) {
	daoPaiIndex.Lock.Lock()
	defer func() {
		daoPaiIndex.Lock.Unlock()
	}()
	daoPaiIndex.DaoPaiFile.Seek(int64(index), 0)
	// var wordbytes []byte = make([]byte, 0)
	var cache []byte = make([]byte, 1024)
	for {
		n, err := daoPaiIndex.DaoPaiFile.Read(cache)
		if err != nil {
			return "", errors.New("not found")
		}
		at := strings.Index(string(cache[:n]), ":")
		if at != -1 {
			return string(cache[:at]), nil
		}
	}
	return "", nil
}

func (daoPaiIndex *DaoPaiIndex) GetDaoPaiListByIndex(index int) ([]*DaoPaiElment, error) {
	result := make([]*DaoPaiElment, 0)
	daoPaiIndex.Lock.Lock()
	defer func() {
		daoPaiIndex.Lock.Unlock()
	}()
	daoPaiIndex.DaoPaiFile.Seek(int64(index), 0)
	cache := make([]byte, 1024)
	n, err := daoPaiIndex.DaoPaiFile.Read(cache)
	for i := 0; i < n; i++ {
		if cache[i] == ':' { // mysql:2fd
			size := utils.BytesToInt(cache[i+1 : i+9])
			daoPaiIndex.DaoPaiFile.Seek(int64(index)+int64(i+1+8), 0)
			_cache := make([]byte, size)

			daoPaiIndex.DaoPaiFile.Read(_cache)
			// fmt.Println()
			// fmt.Println(_cache, size)
			var in bytes.Buffer
			var out bytes.Buffer
			in.Write(_cache)
			r, _err := zlib.NewReader(&in)
			if _err != nil {
				fmt.Println(_err)
			}
			io.Copy(&out, r)
			err = json.Unmarshal(out.Bytes(), &result)
			return result, err
		}
	}
	return []*DaoPaiElment{}, err
}
