package structure

import (
	"sync"
)

const (

	// 判断机器 32 还是 64

	ptrBits     = 32 << uint(^uintptr(0)>>63) //创建指针
	ptrModMask  = ptrBits - 1                 //n&31   ==n%32     n%63 ==n%64
	ptrshift    = (1<<7 + ptrBits) >> 5
	byteModMask = 7 //2^3-1
	byteShift   = 3 //2^3
)

type BitSet interface {
	Get(i int) bool        //提取数据
	Set(i int)             //设置数据
	Unset(i int)           //取消设置
	SetBool(i int, b bool) //设置
}

type Pointers []uintptr

//插入数据移动
func NewPointers(numBits int) Pointers {
	return make(Pointers, (numBits+ptrModMask)>>ptrshift) // ptrshift=6
}

//提取数据
func (p Pointers) Get(i int) bool { // 是否存在
	// return p[uint(i>>ptrshift]&(1<<(uint(i)&ptrModMask)) != 0
	return p[i>>ptrshift]&(1<<(i&ptrModMask)) != 0
}

//设置
func (p Pointers) Set(i int) {
	// 0000 8000
	if i>>ptrshift > len(p) {
		p.Grow(16)
	}
	// p[uint(i)>>ptrshift] |= 1 << (uint(i) & ptrModMask) //设置
	p[i>>ptrshift] |= 1 << (i & ptrModMask) //设置
}

//取消设置
func (p Pointers) UnSet(i int) {
	// p[uint(i)>>ptrshift] &^= 1 << (uint(i) & ptrModMask) //设置
	p[i>>ptrshift] &^= 1 << (i & ptrModMask) //设置
	// p[uint(i)>>ptrshift] = A
	// 1<<(uint(i)&ptrModMask) = B
	// C = A ^ B
	// A = C & A
}

//设置好了
func (p Pointers) Setbool(i int, b bool) {
	if b {
		p.Set(i)
		return
	}
	p.UnSet(i) //取消设置
}

//增长数据,开辟存储空间
func (p *Pointers) Grow(numBits int) {
	ptrs := *p //取出内容
	targetlen := (numBits + ptrModMask) >> ptrshift
	missing := targetlen - len(ptrs)
	if missing > 0 && missing <= targetlen {
		*p = append(ptrs, make(Pointers, missing)...)
	}

}

type Item interface {
	Less(than Item) bool
	Equal(then Item) bool
	EqualId(id int) bool
	AddValue(grade int)
	Add(than Item) Item
}

type Heap struct {
	lock     *sync.Mutex
	data     []Item
	min      bool
	pointers Pointers
}

//标准堆
func NewHeap() *Heap {
	return &Heap{&sync.Mutex{}, make([]Item, 0), true, NewPointers(64)}
}

//最小堆
func NewMin() *Heap {
	return &Heap{new(sync.Mutex), make([]Item, 0), true, NewPointers(64)}
}

//最大堆
func NewMax() *Heap {
	return &Heap{new(sync.Mutex), make([]Item, 0), false, NewPointers(64)}
}
func (h *Heap) isEmpty() bool {
	return len(h.data) == 0
}
func (h *Heap) Len() int {
	return len(h.data)
}

func (h *Heap) Get(index int) Item {
	return h.data[index]
}

func (h *Heap) Insert(it Item) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.data = append(h.data, it)
	h.shiftUp()
}
func (h *Heap) AddValue(id int, grade int) {
	h.lock.Lock()
	defer h.lock.Unlock()
	for i := 0; i < len(h.data); i++ {
		if h.data[i].EqualId(id) {
			id = i
			h.data[i].AddValue(grade)
			break
		}
	}

	h.shiftUpCus(id)
}
func (h *Heap) IsExist(id int) bool {
	return h.pointers.Get(id)
}
func (h *Heap) Less(a, b Item) bool {
	if h.min {
		return a.Less(b)
	} else {
		return b.Less(a)
	}
}
func (h *Heap) Equal(a, b Item) bool {
	return a.Equal(b)
}
func (h *Heap) Extract() Item {
	h.lock.Lock()
	defer h.lock.Unlock()
	if h.Len() == 0 {
		return nil
	}
	el := h.data[0]
	last := h.data[h.Len()-1]
	if h.Len() == 1 {
		h.data = nil
		return el
	}
	h.data = append([]Item{last}, h.data[1:h.Len()-1]...)
	h.shiftDown()

	return el
}

//弹出一个极大值
func (h *Heap) shiftUp() {
	// 1
	//2 3
	// 4 5 6 7
	// 将最值(最小或最大)排在第一个
	for i, parent := h.Len()-1, h.Len()-1; i > 0; i = parent {
		parent = i / 2
		if h.Less(h.Get(i), h.Get(parent)) {
			h.data[parent], h.data[i] = h.data[i], h.data[parent]
		} else {
			break
		}
	}
}

func (h *Heap) shiftUpCus(index int) bool {
	// 1
	//2 3
	// 4 5 6 7
	// 将最值(最小或最大)排在第一个
	var _bool bool = false
	for i, parent := index, index; i > 0; i = parent {
		parent = i / 2
		if h.Less(h.Get(i), h.Get(parent)) {
			_bool = true
			h.data[parent], h.data[i] = h.data[i], h.data[parent]
		} else {
			return _bool
		}
	}
	return _bool
}

func (h *Heap) ShiftDownCus(index int) bool {
	var _bool bool = false
	for i, child := index, 1; i < h.Len() && i*2+1 < h.Len(); i = child {
		child = i*2 + 1
		if child+1 <= h.Len()-1 && h.Less(h.Get(child+1), h.Get(child)) {
			child++ // 循环左右节点过程
		}
		if h.Less(h.Get(i), h.Get(child)) {
			return _bool
		}
		h.data[i], h.data[child] = h.data[child], h.data[i] //处理数据交换
	}
	return _bool

}

//弹出一个极小值
func (h *Heap) shiftDown() {
	// 1
	//3 2
	// 4 5 6 7
	// 堆排序循环过程
	for i, child := 0, 1; i < h.Len() && i*2+1 < h.Len(); i = child {
		child = i*2 + 1
		if child+1 <= h.Len()-1 && h.Less(h.Get(child+1), h.Get(child)) {
			child++ // 循环左右节点过程
		}
		if h.Less(h.Get(i), h.Get(child)) {
			break
		}
		h.data[i], h.data[child] = h.data[child], h.data[i] //处理数据交换
	}
}
