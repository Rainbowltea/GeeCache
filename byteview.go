package geecahe

//缓存值的抽象与封装
//抽象一个只读数据结构，用来表示缓存值
type ByteView struct {
	b []byte
}

func (v ByteView) Len() int {
	return len(v.b)
}

//返回格式为切片的数据
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// String returns the data as a string, making a copy if necessary.
func (v ByteView) String() string {
	return string(v.b)
}

//防止缓存值被外部程序修改。
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}
