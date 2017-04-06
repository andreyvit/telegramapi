package tl

type Object interface {
	Cmd() uint32
	ReadBareFrom(r *Reader)
	WriteBareTo(w *Writer)
}
