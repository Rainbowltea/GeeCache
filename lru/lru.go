package lru

import "container/list"

//LRU算法   map +  双向链表队列
//最近最少使用，最近使用的数据移动到队尾
type Cache struct {
	maxBytes  int64 //允许使用的最大内存
	nbytes    int64 //当前已经使用的内存
	ll        *list.List
	cache     map[string]*list.Element      //字典定义：键值是字符串，值是双向链表中对应节点的指针
	OnEvicted func(key string, value Value) //某条记录被移除时的回调函数
}

//双向链表节点的数据类型，淘汰队首节点时，需要"key"从字典中删除对应的映射
type entry struct {
	key   string
	value Value
}

//为了通用性，值是实现了Value接口的任意类型，该接口只包含一个方法用来计算占用多少字节
type Value interface {
	Len() int
}

// New is the Constructor of Cache
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

//查找功能
//1，从字典中找到对应的双向链表的节点
//2，将该节点移动到队尾（自定义为  队尾==front）
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok != false {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}
	return
}

//删除
//缓存淘汰，即为移除最近最少访问的节点=队首
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

//新增和修改
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	//若空间不够循环腾出已用空间给新增数据
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// Len the number of cache entries
func (c *Cache) Len() int {
	return c.ll.Len()
}
