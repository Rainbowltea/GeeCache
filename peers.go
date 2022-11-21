package geecahe

// PeerPicker is the interface that must be implemented to locate
// the peer that owns a specific key.
//选节点
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

//从对应的group查找缓存值
// PeerGetter is the interface that must be implemented by a peer.
type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
}
