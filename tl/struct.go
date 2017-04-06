package tl

type Struct interface {
	Cmd() uint32
	ReadFrom(r *Reader)
	WriteTo(w *Writer)
}
