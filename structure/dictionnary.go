package structure

import (
	"errors"
	"strings"
)

type LruElement struct {
	Key   string
	Value interface{}
}
type LRU struct {
	K        int
	Elements []LruElement
}
type WordNode struct {
	Word  string
	Index int
	List  []*DaoPaiElment
	Left  *WordNode
	Right *WordNode
}
type WordTree struct {
	Root *WordNode
	Lru  *LRU
}

func (lru *LRU) Get(key string) (interface{}, error) {
	for i := 0; i < len(lru.Elements); i++ {
		if lru.Elements[i].Key == key {
			newlru := make([]LruElement, 0)
			newlru = append(newlru, lru.Elements[i])
			newlru = append(newlru, lru.Elements[:i]...)
			if i != len(lru.Elements)-1 {
				newlru = append(newlru, lru.Elements[i+1:]...)
			}
			return lru.Elements[i].Value, nil
		}
	}
	return 0, errors.New("not found")
}
func (lru *LRU) Insert(key string, value interface{}) {
	if len(lru.Elements) < lru.K {
		lru.Elements = append(lru.Elements, LruElement{
			Key:   key,
			Value: value,
		})
	} else {
		for i := len(lru.Elements) - 1; i > 0; i-- {
			lru.Elements[i].Key, lru.Elements[i].Value = lru.Elements[i-1].Key, lru.Elements[i-1].Value
		}
		lru.Elements[0] = LruElement{
			Key:   key,
			Value: value,
		}
	}
}
func NewLru(k int) *LRU {
	return &LRU{K: k, Elements: make([]LruElement, 0)}
}
func NewIndexWordNode(word string, index int) *WordNode {
	return &WordNode{Word: word, Index: index, Left: nil, Right: nil}
}
func NewListWordNode(word string, data *DaoPaiElment) *WordNode {
	return &WordNode{Word: word, List: []*DaoPaiElment{data}, Left: nil, Right: nil}
}
func NewWordTree(k int) *WordTree {
	return &WordTree{Lru: NewLru(k), Root: nil}
}
func (wordtree *WordTree) InsertIndexNode(word string, index int) {
	if wordtree.Root == nil {
		wordtree.Root = NewIndexWordNode(word, index)
		wordtree.Lru.Insert(word, index)
	} else {
		wordtree.Root.insertIndex(NewIndexWordNode(word, index))
		wordtree.Lru.Insert(word, index)
	}
}
func (wordtree *WordTree) InsertListNode(word string, data *DaoPaiElment) {
	if wordtree.Root == nil {
		wordtree.Root = NewListWordNode(word, data)
	} else {
		wordtree.Root.insertList(NewListWordNode(word, data))
	}
}

// 根据 单词 查找索引
func (wordtree *WordTree) GetIndex(word string) (int, error) {
	if word == "" {
		return 0, errors.New("arguments position 1 is empty string")
	}
	// var index interface{}
	// var err error

	// 从LRU中找
	// index, err = wordtree.Lru.Get(word)
	// if err == nil {
	// 	return index.(int), nil
	// }
	node, _err := wordtree.Search(word)
	if _err != nil {
		return 0, _err
	} else {
		wordtree.Lru.Insert(word, node.Index)
		return node.Index, nil
	}

}
func (wordtree *WordTree) Traverse() []WordNode {
	if wordtree.Root == nil {
		return []WordNode{}
	}
	datas := []WordNode{}
	wordtree.Root.Traverse(&datas)
	return datas
}
func (wordnode *WordNode) Traverse(datas *[]WordNode) {
	if wordnode.Left != nil {
		wordnode.Left.Traverse(datas)
	}
	*datas = append(*datas, *wordnode)
	if wordnode.Right != nil {
		wordnode.Right.Traverse(datas)
	}
}
func (wordtree *WordTree) GetList(word string) ([]*DaoPaiElment, error) {
	if word == "" {
		return []*DaoPaiElment{}, errors.New("arguments position 1 is empty string")
	}

	var err error

	node, err := wordtree.Search(word)
	if err != nil {
		return []*DaoPaiElment{}, err
	} else {
		return node.List, nil
	}

}

func (wordTree *WordTree) Search(word string) (*WordNode, error) {
	if wordTree.Root == nil {
		return &WordNode{}, errors.New("not found")
	}
	return wordTree.Root.search(word)
}
func (wordNode *WordNode) search(word string) (*WordNode, error) {
	if wordNode == nil {
		return &WordNode{}, errors.New("not found")
	} else {
		if strings.ToLower(wordNode.Word) < strings.ToLower(word) {
			return wordNode.Right.search(word)
		} else if wordNode.Word > word {
			return wordNode.Left.search(word)
		}
	}
	return wordNode, nil

}

func (wordnode *WordNode) insertList(node *WordNode) {

	if node == nil {
		return
	}
	if wordnode.Word > node.Word {
		if wordnode.Left == nil {
			wordnode.Left = node
		} else {
			wordnode.Left.insertList(node)
		}
	} else if wordnode.Word < node.Word {
		if wordnode.Right == nil {
			wordnode.Right = node
		} else {
			wordnode.Right.insertList(node)
		}
	} else {
		wordnode.List = append(wordnode.List, node.List...)
	}
}

func (wordnode *WordNode) insertIndex(node *WordNode) {
	if node == nil {
		return
	}
	if wordnode.Word > node.Word {
		if wordnode.Left == nil {
			wordnode.Left = node
		} else {
			wordnode.Left.insertIndex(node)
		}
	} else if wordnode.Word < node.Word {
		if wordnode.Right == nil {
			wordnode.Right = node
		} else {
			wordnode.Right.insertIndex(node)
		}
	}
}
