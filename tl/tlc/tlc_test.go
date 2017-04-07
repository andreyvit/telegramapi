package tlc

import (
	"github.com/andreyvit/diff"
	"github.com/andreyvit/telegramapi/tl/knownschemas"
	"github.com/andreyvit/telegramapi/tl/tlschema"
	"testing"
)

func TestSimple(t *testing.T) {
	sch := tlschema.MustParse(`
        nearestDc#8e1a1775 country:string this_dc:int nearest_dc:int = NearestDc;
        --- functions ---
        help.getNearestDc#1fb33026 = NearestDc;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        // TLNearestDc represents nearestDc from TL schema
        type TLNearestDc struct {
            Country   string // country: string
            ThisDc    int    // this_dc: int
            NearestDc int    // nearest_dc: int
        }

        func (s *TLNearestDc) Cmd() uint32 {
            return TagNearestDc
        }

        func (s *TLNearestDc) ReadBareFrom(r *tl.Reader) {
            s.Country = r.ReadString()
            s.ThisDc = r.ReadInt()
            s.NearestDc = r.ReadInt()
        }

        func (s *TLNearestDc) WriteBareTo(w *tl.Writer) {
            w.WriteString(s.Country)
            w.WriteInt(s.ThisDc)
            w.WriteInt(s.NearestDc)
        }

        // TLHelpGetNearestDc represents help.getNearestDc from TL schema
        type TLHelpGetNearestDc struct {
        }

        func (s *TLHelpGetNearestDc) Cmd() uint32 {
            return TagHelpGetNearestDc
        }

        func (s *TLHelpGetNearestDc) ReadBareFrom(r *tl.Reader) {
        }

        func (s *TLHelpGetNearestDc) WriteBareTo(w *tl.Writer) {
        }

        func ReadBoxedObjectFrom(r *tl.Reader) tl.Object {
            cmd := r.ReadCmd()
            switch cmd {
            case TagNearestDc:
                s := new(TLNearestDc)
                s.ReadBareFrom(r)
                return s
            case TagHelpGetNearestDc:
                s := new(TLHelpGetNearestDc)
                s.ReadBareFrom(r)
                return s
            default:
                return nil
            }
        }
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestInt(t *testing.T) {
	sch := tlschema.MustParse(`
        foo#11223344 bar:int = Foo;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        // TLFoo represents foo from TL schema
        type TLFoo struct {
            Bar int // bar: int
        }

        func (s *TLFoo) Cmd() uint32 {
            return TagFoo
        }

        func (s *TLFoo) ReadBareFrom(r *tl.Reader) {
            s.Bar = r.ReadInt()
        }

        func (s *TLFoo) WriteBareTo(w *tl.Writer) {
            w.WriteInt(s.Bar)
        }
        
        func ReadBoxedObjectFrom(r *tl.Reader) tl.Object {
            cmd := r.ReadCmd()
            switch cmd {
            case TagFoo:
                s := new(TLFoo)
                s.ReadBareFrom(r)
                return s
            default:
                return nil
            }
        }
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestBigInt(t *testing.T) {
	sch := tlschema.MustParse(`
        resPQ#11223344 pq:bytes = ResPQ;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        // TLResPQ represents resPQ from TL schema
        type TLResPQ struct {
            PQ *big.Int // pq: bytes
        }

        func (s *TLResPQ) Cmd() uint32 {
            return TagResPQ
        }

        func (s *TLResPQ) ReadBareFrom(r *tl.Reader) {
            s.PQ = r.ReadBigInt()
        }

        func (s *TLResPQ) WriteBareTo(w *tl.Writer) {
            w.WriteBigInt(s.PQ)
        }
        
        func ReadBoxedObjectFrom(r *tl.Reader) tl.Object {
            cmd := r.ReadCmd()
            switch cmd {
            case TagResPQ:
                s := new(TLResPQ)
                s.ReadBareFrom(r)
                return s
            default:
                return nil
            }
        }
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestVectorBareInt(t *testing.T) {
	sch := tlschema.MustParse(`
        foo#11223344 bar:Vector<int> = Foo;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        // TLFoo represents foo from TL schema
        type TLFoo struct {
            Bar []int // bar: Vector<int>
        }

        func (s *TLFoo) Cmd() uint32 {
            return TagFoo
        }

        func (s *TLFoo) ReadBareFrom(r *tl.Reader) {
            if cmd := r.ReadCmd(); cmd != TagVector {
                r.Fail(errors.New("expected: vector"))
            }
            s.Bar = make([]int, r.ReadInt())
            for i := 0; i < len(s.Bar); i++ {
                s.Bar[i] = r.ReadInt()
            }
        }

        func (s *TLFoo) WriteBareTo(w *tl.Writer) {
            w.WriteCmd(TagVector)
            w.WriteInt(len(s.Bar))
            for i := 0; i < len(s.Bar); i++ {
                w.WriteInt(s.Bar[i])
            }
        }        

        func ReadBoxedObjectFrom(r *tl.Reader) tl.Object {
            cmd := r.ReadCmd()
            switch cmd {
            case TagFoo:
                s := new(TLFoo)
                s.ReadBareFrom(r)
                return s
            default:
                return nil
            }
        }
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestBareVectorBareInt(t *testing.T) {
	sch := tlschema.MustParse(`
        foo#11223344 bar:%Vector<int> = Foo;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        // TLFoo represents foo from TL schema
        type TLFoo struct {
            Bar []int // bar: %Vector<int>
        }

        func (s *TLFoo) Cmd() uint32 {
            return TagFoo
        }

        func (s *TLFoo) ReadBareFrom(r *tl.Reader) {
            s.Bar = make([]int, r.ReadInt())
            for i := 0; i < len(s.Bar); i++ {
                s.Bar[i] = r.ReadInt()
            }
        }

        func (s *TLFoo) WriteBareTo(w *tl.Writer) {
            w.WriteInt(len(s.Bar))
            for i := 0; i < len(s.Bar); i++ {
                w.WriteInt(s.Bar[i])
            }
        }        

        func ReadBoxedObjectFrom(r *tl.Reader) tl.Object {
            cmd := r.ReadCmd()
            switch cmd {
            case TagFoo:
                s := new(TLFoo)
                s.ReadBareFrom(r)
                return s
            default:
                return nil
            }
        }
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestBareVectorBareStruct(t *testing.T) {
	sch := tlschema.MustParse(`
        foo#11223344 bar:vector<%Boz> = Foo;
        boz#99887766 = Boz;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        // TLFoo represents foo from TL schema
        type TLFoo struct {
            Bar []*TLBoz // bar: vector<%Boz>
        }

        func (s *TLFoo) Cmd() uint32 {
            return TagFoo
        }

        func (s *TLFoo) ReadBareFrom(r *tl.Reader) {
            s.Bar = make([]*TLBoz, r.ReadInt())
            for i := 0; i < len(s.Bar); i++ {
                s.Bar[i] = new(TLBoz)
                s.Bar[i].ReadBareFrom(r)
            }
        }

        func (s *TLFoo) WriteBareTo(w *tl.Writer) {
            w.WriteInt(len(s.Bar))
            for i := 0; i < len(s.Bar); i++ {
                s.Bar[i].WriteBareTo(w)
            }
        }   

        // TLBoz represents boz from TL schema
        type TLBoz struct {
        }

        func (s *TLBoz) Cmd() uint32 {
            return TagBoz
        }

        func (s *TLBoz) ReadBareFrom(r *tl.Reader) {
        }

        func (s *TLBoz) WriteBareTo(w *tl.Writer) {
        }

        func ReadBoxedObjectFrom(r *tl.Reader) tl.Object {
            cmd := r.ReadCmd()
            switch cmd {
            case TagFoo:
                s := new(TLFoo)
                s.ReadBareFrom(r)
                return s
            case TagBoz:
                s := new(TLBoz)
                s.ReadBareFrom(r)
                return s
            default:
                return nil
            }
        }
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestBareVectorBoxedStruct(t *testing.T) {
	sch := tlschema.MustParse(`
        foo#11223344 bar:vector<Boz> = Foo;
        boz#99887766 = Boz;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        // TLFoo represents foo from TL schema
        type TLFoo struct {
            Bar []*TLBoz // bar: vector<Boz>
        }

        func (s *TLFoo) Cmd() uint32 {
            return TagFoo
        }

        func (s *TLFoo) ReadBareFrom(r *tl.Reader) {
            s.Bar = make([]*TLBoz, r.ReadInt())
            for i := 0; i < len(s.Bar); i++ {
                if cmd := r.ReadCmd(); cmd != TagBoz {
                    r.Fail(errors.New("expected: boz"))
                }
                s.Bar[i] = new(TLBoz)
                s.Bar[i].ReadBareFrom(r)
            }
        }

        func (s *TLFoo) WriteBareTo(w *tl.Writer) {
            w.WriteInt(len(s.Bar))
            for i := 0; i < len(s.Bar); i++ {
                w.WriteCmd(TagBoz)
                s.Bar[i].WriteBareTo(w)
            }
        }   

        // TLBoz represents boz from TL schema
        type TLBoz struct {
        }

        func (s *TLBoz) Cmd() uint32 {
            return TagBoz
        }

        func (s *TLBoz) ReadBareFrom(r *tl.Reader) {
        }

        func (s *TLBoz) WriteBareTo(w *tl.Writer) {
        }

        func ReadBoxedObjectFrom(r *tl.Reader) tl.Object {
            cmd := r.ReadCmd()
            switch cmd {
            case TagFoo:
                s := new(TLFoo)
                s.ReadBareFrom(r)
                return s
            case TagBoz:
                s := new(TLBoz)
                s.ReadBareFrom(r)
                return s
            default:
                return nil
            }
        }
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestMultiCtorType(t *testing.T) {
	sch := tlschema.MustParse(`
        foo#11223344 x:int = Foo;
        bar#99887766 y:string = Foo;
    `)
	code := GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
	expected := `
        // TLFooType represents Foo from TL schema
        type TLFooType interface {
            IsTLFoo()
            Cmd() uint32
            ReadBareFrom(r *tl.Reader)
            WriteBareTo(w *tl.Writer)
        }

        // TLFoo represents foo from TL schema
        type TLFoo struct {
            X int // x: int
        }

        func (s *TLFoo) IsTLFoo() {}

        func (s *TLFoo) Cmd() uint32 {
            return TagFoo
        }

        func (s *TLFoo) ReadBareFrom(r *tl.Reader) {
            s.X = r.ReadInt()
        }

        func (s *TLFoo) WriteBareTo(w *tl.Writer) {
            w.WriteInt(s.X)
        }

        // TLBar represents bar from TL schema
        type TLBar struct {
            Y string // y: string
        }

        func (s *TLBar) IsTLFoo() {}

        func (s *TLBar) Cmd() uint32 {
            return TagBar
        }

        func (s *TLBar) ReadBareFrom(r *tl.Reader) {
            s.Y = r.ReadString()
        }

        func (s *TLBar) WriteBareTo(w *tl.Writer) {
            w.WriteString(s.Y)
        }

        func ReadBoxedObjectFrom(r *tl.Reader) tl.Object {
            cmd := r.ReadCmd()
            switch cmd {
            case TagFoo:
                s := new(TLFoo)
                s.ReadBareFrom(r)
                return s
            case TagBar:
                s := new(TLBar)
                s.ReadBareFrom(r)
                return s
            default:
                return nil
            }
        }       
    `
	a, e := diff.TrimLinesInString(code), diff.TrimLinesInString(expected)
	if a != e {
		t.Errorf("Code not as expected:\n%v\n\nActual:\n%s", diff.LineDiff(e, a), code)
	}
}

func TestMTProto(t *testing.T) {
	sch := tlschema.MustParse(knownschemas.MTProtoSchema)
	GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
}

func TestTelegram(t *testing.T) {
	sch := tlschema.MustParse(knownschemas.TelegramSchema)
	GenerateGoCode(sch, Options{PackageName: "foo", SkipPrelude: true})
}
