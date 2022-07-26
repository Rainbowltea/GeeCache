package GeeCache

import (
	"GeeCache/singleflight"
	"fmt"
	"log"
	"sync"
)

type Group struct {
	name      string
	getter    Getter
	mainCache cache
	peers     PeerPicker
	loader    *singleflight.Group
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

//将group存储在全局变量groups中
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (value ByteView, err error) {

	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	return
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err

	}
	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

//定义接口和回调函数
type Getter interface {
	Get(key string) ([]byte, error)
}

//实现Getter接口的方法
//接口型函数：函数类型实现某一个接口，方便使用者在调用时既能够传入函数作为参数，
//也能够传入实现了该接口的结构体作为参数。
//定义一个函数类型 F，并且实现接口 A 的方法，然后在这个方法中调用自己。这是 Go
//语言中将其他函数（参数返回值定义与 F 一致）转换为接口 A 的常用技巧。
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

//定义函数类型
type GetterFunc func(key string) ([]byte, error)

//
//借助GetterFunc的类型转化，将一个匿名回调函数转换成了接口f Getter
// func TestGetter(t *testing.T) {
// 	var f Getter = GetterFunc(func(key string) ([]byte, error) {
// 		return []byte(key), nil
// 	})

// 	expect := []byte("key")
// 	if v, _ := f.Get("key"); !reflect.DeepEqual(v, expect) {
// 		t.Errorf("callback failed")
// 	}
// }

// RegisterPeers registers a PeerPicker for choosing remote peer
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}
