package main

import (
	"bufio"
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"go_code/search_engine/parse"
	"go_code/search_engine/structure"
	"go_code/search_engine/utils"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// 老索引
var OldIndex *structure.DaoPaiIndex
var rootpath string = "./"

// // 临时索引
var NewIndex *structure.DaoPaiIndex

type pileElement struct {
	DocumentId int // 文档id
	Grade      int // 分数
}

func (x *pileElement) Less(than structure.Item) bool {
	return x.Grade < than.(*pileElement).Grade
}
func (x *pileElement) Equal(than structure.Item) bool {
	return x.DocumentId == than.(*pileElement).DocumentId
}
func (x *pileElement) EqualId(id int) bool {
	return x.DocumentId == id
}
func (x *pileElement) AddValue(grade int) {
	x.Grade = x.Grade + grade
}
func (x *pileElement) Add(than structure.Item) structure.Item {
	return &pileElement{
		x.DocumentId,
		x.Grade + than.(*pileElement).Grade,
	}
}
func init() {
	NewIndex = &structure.DaoPaiIndex{Dictionary: structure.NewWordTree(50), Lock: &sync.Mutex{}, IsDisk: false}
	OldIndex = &structure.DaoPaiIndex{Dictionary: structure.NewWordTree(10), Lock: &sync.Mutex{}, IsDisk: true}
	// OldIndex.LoadDaoPaiFile("./data/1.txt")
	file, err := os.Open(filepath.Join(rootpath, "data", "current.txt"))
	if err != nil {
		panic(err)
	}
	reader := bufio.NewReader(file)
	data, _, _ := reader.ReadLine()
	indexfilename := strings.TrimSpace(string(data))
	OldIndex.LoadDaoPaiFile(filepath.Join(rootpath, "data", indexfilename))

}
func SaveToDisk() {
	now := time.Now().Unix()
	_file, err := os.OpenFile(filepath.Join(rootpath, "data", "current.txt"), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0655)
	if err != nil {
		panic(err)
	}
	_file.Write([]byte(fmt.Sprintf("%d.txt", now)))
	newFileName := fmt.Sprintf("%s/%d.txt", filepath.Join(rootpath, "data"), now)
	newDaopaiIndex := &structure.DaoPaiIndex{Dictionary: structure.NewWordTree(50), IsDisk: true}
	newDaopaiIndex.DaoPaiFile, _ = os.OpenFile(newFileName, os.O_CREATE|os.O_RDWR, 0655)
	Olds := OldIndex.Dictionary.Traverse()
	News := NewIndex.Dictionary.Traverse()
	i := 0
	j := 0
	for i < len(Olds) && j < len(News) {
		if Olds[i].Word < News[j].Word {
			frontByte := []byte(Olds[i].Word + ":")
			var in bytes.Buffer
			data, err := OldIndex.GetDaoPaiListByIndex(Olds[i].Index)
			if err != nil {
				fmt.Println(err)
				i++
				continue
			}
			databyte, _err := json.Marshal(data)
			if _err != nil {
				fmt.Println(_err)
				i++
				continue
			}

			w, _ := zlib.NewWriterLevel(&in, zlib.BestCompression)
			w.Write(databyte)
			w.Close()
			size := len(in.Bytes())
			sizebyte := utils.IntTobytes(int64(size))
			frontByte = append(frontByte, sizebyte...)
			newDaopaiIndex.DaoPaiFile.Write(append(frontByte, in.Bytes()...))
			i++
		} else if Olds[i].Word > News[j].Word {
			var in bytes.Buffer
			frontByte := []byte(News[j].Word + ":")
			data := News[j].List
			databyte, err := json.Marshal(data)
			if err != nil {
				fmt.Println(err)
				j++
				continue
			}

			w, _ := zlib.NewWriterLevel(&in, zlib.BestCompression)
			w.Write(databyte)
			w.Close()
			size := len(in.Bytes())
			sizebyte := utils.IntTobytes(int64(size))
			frontByte = append(frontByte, sizebyte...)
			newDaopaiIndex.DaoPaiFile.Write(append(frontByte, in.Bytes()...))
			j++
		} else {
			// 相等
			frontByte := []byte(Olds[i].Word + ":")
			var in bytes.Buffer
			olddata, err := OldIndex.GetDaoPaiListByIndex(Olds[i].Index)
			if err != nil {
				fmt.Println(err)

			}
			newdata := append(News[j].List, olddata...)

			databyte, err := json.Marshal(newdata)
			if err != nil {
				fmt.Println(err)
				j++
				i++
				continue
			}

			w, _ := zlib.NewWriterLevel(&in, zlib.BestCompression)
			w.Write(databyte)
			w.Close()

			size := len(in.Bytes())
			sizebyte := utils.IntTobytes(int64(size))
			frontByte = append(frontByte, sizebyte...)
			newDaopaiIndex.DaoPaiFile.Write(append(frontByte, in.Bytes()...))
			i++
			j++
		}
	}
	if i < len(Olds) {
		for p := i; p < len(Olds); p++ {
			frontByte := []byte(Olds[p].Word + ":")
			var in bytes.Buffer
			data, err := OldIndex.GetDaoPaiListByIndex(Olds[p].Index)
			if err != nil {
				fmt.Println(err)
				continue
			}
			databyte, _err := json.Marshal(data)
			if _err != nil {
				fmt.Println(_err)
				continue
			}

			w, _ := zlib.NewWriterLevel(&in, zlib.BestCompression)
			w.Write(databyte)
			w.Close()

			size := len(in.Bytes())
			sizebyte := utils.IntTobytes(int64(size))
			frontByte = append(frontByte, sizebyte...)
			newDaopaiIndex.DaoPaiFile.Write(append(frontByte, in.Bytes()...))
		}
	} else {
		for p := j; p < len(News); p++ {
			var in bytes.Buffer
			frontByte := []byte(News[p].Word + ":")
			data := News[p].List
			databyte, err := json.Marshal(data)
			if err != nil {
				fmt.Println(err)
				continue
			}

			w, _ := zlib.NewWriterLevel(&in, zlib.BestCompression)
			w.Write(databyte)
			w.Close()
			size := len(in.Bytes())
			sizebyte := utils.IntTobytes(int64(size))
			frontByte = append(frontByte, sizebyte...)
			fmt.Println(in.Bytes())
			newDaopaiIndex.DaoPaiFile.Write(append(frontByte, in.Bytes()...))
		}
	}
	newDaopaiIndex.DaoPaiFile.Close()
	newDaopaiIndex.LoadDaoPaiFile(newFileName)
	OldIndex = newDaopaiIndex
	OldIndex.Lock = &sync.Mutex{}
	NewIndex = &structure.DaoPaiIndex{Dictionary: structure.NewWordTree(50), Lock: &sync.Mutex{}, IsDisk: false}
}

func main() {

	// time.Sleep(time.Second * 1000)
	var title string
	var content string
	var id int
	for {
		fmt.Scanf("%d  %s %s\n", &id, &title, &content)
		if title == "stop" {
			break
		}
		if title == "save" {
			SaveToDisk()
			// break
			continue
		}
		if title == "parse" {
			wordmap1 := parse.ParseString(content)
			for p, v := range wordmap1 {
				fmt.Println(p, v)
			}
			continue
		}
		if title == "search" {
			maps := parse.ParseString(content)
			status := make(map[int]bool, 0)
			pile := structure.NewMax()
			for word, _ := range maps {
				data1, err1 := OldIndex.GetList(word)
				if err1 != nil && err1.Error() != "not found" {
					panic(err1)
				}
				data2, err2 := NewIndex.GetList(word)
				if err2 != nil && err2.Error() != "not found" {
					panic(err2)
				}
				fmt.Println(len(data1), len(data2))
				data := append(data1, data2...)
				grade := 0
				for _, item := range data {
					if item.IsTitle == 1 {
						grade += 2
					}
					grade += item.TF
					_, ok := status[item.DocumentId]
					if ok {
						pile.AddValue(item.DocumentId, grade)
					} else {
						status[item.DocumentId] = true
						pile.Insert(&pileElement{item.DocumentId, grade})
					}
					// fmt.Printf("Id:%d  TF:%d  出现在Title: %d\n", item.DocumentId, item.TF, item.IsTitle)
					grade = 0
				}
			}

			for {
				el := pile.Extract()
				if el == nil {
					break
				}
				fmt.Println(el.(*pileElement))
			}
			// SaveToDisk()
			// break
			continue
		}
		fmt.Println(id, title, content)
		flag := NewIndex.ParseArticle(id, title, content)
		if flag {
			// SaveToDisk()
		}
	}

	// arr := OldIndex.Dictionary.Traverse()
	// for _, v := range arr {
	// 	fmt.Printf("===== %s: =======", v.Word)
	// 	data, _ := OldIndex.GetDaoPaiListByIndex(v.Index)
	// 	for _, item := range data {
	// 		fmt.Printf("Id:%d  TF:%d  出现在Title: %d\n", item.DocumentId, item.TF, item.IsTitle)

	// 	}
	// }

	// 从磁盘中载入索引

	// data := structure.DaoPaiElment{1, 2, 3}
	// var in bytes.Buffer
	// b, _ := json.Marshal(data)
	// fmt.Println(len(b))
	// w, _ := zlib.NewWriterLevel(&in, zlib.BestCompression)
	// w.Write(b)
	// w.Close()
	// fmt.Println(len(in.Bytes()), string(in.Bytes()))

	// var out bytes.Buffer
	// r, _ := zlib.NewReader(&in)
	// io.Copy(&out, r)
	// fmt.Println(out.String())
}
