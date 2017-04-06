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
	code := GenerateGoCode(sch, Options{PackageName: "foo"})
	expected := `
        package foo

        import (
            "github.com/andreyvit/telegramapi/tl"
            "math/big"
            "time"
        )

        const (
            TagNearestDc        uint32 = 0x8e1a1775
            TagHelpGetNearestDc        = 0x1fb33026
            TagVector                  = 0x1cb5c415
        )

        // NearestDc represents nearestDc from TL schema
        type NearestDc struct {
            Country   string // country: string
            ThisDc    int    // this_dc: int
            NearestDc int    // nearest_dc: int
        }

        func (s *NearestDc) Cmd() uint32 {
            return TagNearestDc
        }

        func (s *NearestDc) ReadBareFrom(r *tlschema.Reader) {
            s.Country = r.ReadString()
            s.ThisDc = r.ReadInt()
            s.NearestDc = r.ReadInt()
        }

        func (s *NearestDc) WriteBareTo(w *tlschema.Writer) {
            w.WriteString(s.Country)
            w.WriteInt(s.ThisDc)
            w.WriteInt(s.NearestDc)
        }

        // HelpGetNearestDc represents help.getNearestDc from TL schema
        type HelpGetNearestDc struct {
        }

        func (s *HelpGetNearestDc) Cmd() uint32 {
            return TagHelpGetNearestDc
        }

        func (s *HelpGetNearestDc) ReadBareFrom(r *tlschema.Reader) {
        }

        func (s *HelpGetNearestDc) WriteBareTo(w *tlschema.Writer) {
        }

        func ReadBoxedObjectFrom(r *tlschema.Reader) tl.Object {
            cmd := r.ReadCmd()
            switch cmd {
            case TagNearestDc:
                s := new(NearestDc)
                s.ReadBareFrom(r)
                return s
            case TagHelpGetNearestDc:
                s := new(HelpGetNearestDc)
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
	code := GenerateGoCode(sch, Options{PackageName: "foo"})
	expected := `
        package foo

        import (
            "github.com/andreyvit/telegramapi/tl"
            "math/big"
            "time"
        )

        const (
            TagFoo    uint32 = 0x11223344
            TagVector        = 0x1cb5c415
        )

        // Foo represents foo from TL schema
        type Foo struct {
            Bar int // bar: int
        }

        func (s *Foo) Cmd() uint32 {
            return TagFoo
        }

        func (s *Foo) ReadBareFrom(r *tlschema.Reader) {
            s.Bar = r.ReadInt()
        }

        func (s *Foo) WriteBareTo(w *tlschema.Writer) {
            w.WriteInt(s.Bar)
        }
        
        func ReadBoxedObjectFrom(r *tlschema.Reader) tl.Object {
            cmd := r.ReadCmd()
            switch cmd {
            case TagFoo:
                s := new(Foo)
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
	code := GenerateGoCode(sch, Options{PackageName: "foo"})
	expected := `
        package foo

        import (
            "github.com/andreyvit/telegramapi/tl"
            "math/big"
            "time"
        )

        const (
            TagResPQ  uint32 = 0x11223344
            TagVector        = 0x1cb5c415
        )

        // ResPQ represents resPQ from TL schema
        type ResPQ struct {
            Pq *big.Int // pq: bytes
        }

        func (s *ResPQ) Cmd() uint32 {
            return TagResPQ
        }

        func (s *ResPQ) ReadBareFrom(r *tlschema.Reader) {
            s.Pq = r.ReadBigInt()
        }

        func (s *ResPQ) WriteBareTo(w *tlschema.Writer) {
            w.WriteBigInt(s.Pq)
        }
        
        func ReadBoxedObjectFrom(r *tlschema.Reader) tl.Object {
            cmd := r.ReadCmd()
            switch cmd {
            case TagResPQ:
                s := new(ResPQ)
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
	code := GenerateGoCode(sch, Options{PackageName: "foo"})
	expected := `
        package foo

        import (
            "github.com/andreyvit/telegramapi/tl"
            "math/big"
            "time"
        )

        const (
            TagFoo    uint32 = 0x11223344
            TagVector        = 0x1cb5c415
        )

        // Foo represents foo from TL schema
        type Foo struct {
            Bar []int // bar: Vector<int>
        }

        func (s *Foo) Cmd() uint32 {
            return TagFoo
        }

        func (s *Foo) ReadBareFrom(r *tlschema.Reader) {
            if cmd := r.ReadCmd(); cmd != TagVector {
                r.Fail(errors.New("expected: vector"))
            }
            s.Bar = make([]int, r.ReadInt())
            for i := 0; i < len(s.Bar); i++ {
                s.Bar[i] = r.ReadInt()
            }
        }

        func (s *Foo) WriteBareTo(w *tlschema.Writer) {
            w.WriteCmd(TagVector)
            w.WriteInt(len(s.Bar))
            for i := 0; i < len(s.Bar); i++ {
                w.WriteInt(s.Bar[i])
            }
        }        

        func ReadBoxedObjectFrom(r *tlschema.Reader) tl.Object {
            cmd := r.ReadCmd()
            switch cmd {
            case TagFoo:
                s := new(Foo)
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
	code := GenerateGoCode(sch, Options{PackageName: "foo"})
	expected := `
        package foo

        import (
            "github.com/andreyvit/telegramapi/tl"
            "math/big"
            "time"
        )

        const (
            TagFoo    uint32 = 0x11223344
            TagVector        = 0x1cb5c415
        )

        // Foo represents foo from TL schema
        type Foo struct {
            Bar []int // bar: %Vector<int>
        }

        func (s *Foo) Cmd() uint32 {
            return TagFoo
        }

        func (s *Foo) ReadBareFrom(r *tlschema.Reader) {
            s.Bar = make([]int, r.ReadInt())
            for i := 0; i < len(s.Bar); i++ {
                s.Bar[i] = r.ReadInt()
            }
        }

        func (s *Foo) WriteBareTo(w *tlschema.Writer) {
            w.WriteInt(len(s.Bar))
            for i := 0; i < len(s.Bar); i++ {
                w.WriteInt(s.Bar[i])
            }
        }        

        func ReadBoxedObjectFrom(r *tlschema.Reader) tl.Object {
            cmd := r.ReadCmd()
            switch cmd {
            case TagFoo:
                s := new(Foo)
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
	code := GenerateGoCode(sch, Options{PackageName: "foo"})
	expected := `
        package foo

        import (
            "github.com/andreyvit/telegramapi/tl"
            "math/big"
            "time"
        )

        const (
            TagFoo    uint32 = 0x11223344
            TagBoz           = 0x99887766
            TagVector        = 0x1cb5c415
        )

        // Foo represents foo from TL schema
        type Foo struct {
            Bar []*Boz // bar: vector<%Boz>
        }

        func (s *Foo) Cmd() uint32 {
            return TagFoo
        }

        func (s *Foo) ReadBareFrom(r *tlschema.Reader) {
            s.Bar = make([]*Boz, r.ReadInt())
            for i := 0; i < len(s.Bar); i++ {
                s.Bar[i] = new(Boz)
                s.Bar[i].ReadBareFrom(r)
            }
        }

        func (s *Foo) WriteBareTo(w *tlschema.Writer) {
            w.WriteInt(len(s.Bar))
            for i := 0; i < len(s.Bar); i++ {
                s.Bar[i].WriteBareTo(w)
            }
        }   

        // Boz represents boz from TL schema
        type Boz struct {
        }

        func (s *Boz) Cmd() uint32 {
            return TagBoz
        }

        func (s *Boz) ReadBareFrom(r *tlschema.Reader) {
        }

        func (s *Boz) WriteBareTo(w *tlschema.Writer) {
        }

        func ReadBoxedObjectFrom(r *tlschema.Reader) tl.Object {
            cmd := r.ReadCmd()
            switch cmd {
            case TagFoo:
                s := new(Foo)
                s.ReadBareFrom(r)
                return s
            case TagBoz:
                s := new(Boz)
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
	code := GenerateGoCode(sch, Options{PackageName: "foo"})
	expected := `
        package foo

        import (
            "github.com/andreyvit/telegramapi/tl"
            "math/big"
            "time"
        )

        const (
            TagFoo    uint32 = 0x11223344
            TagBoz           = 0x99887766
            TagVector        = 0x1cb5c415
        )

        // Foo represents foo from TL schema
        type Foo struct {
            Bar []*Boz // bar: vector<Boz>
        }

        func (s *Foo) Cmd() uint32 {
            return TagFoo
        }

        func (s *Foo) ReadBareFrom(r *tlschema.Reader) {
            s.Bar = make([]*Boz, r.ReadInt())
            for i := 0; i < len(s.Bar); i++ {
                if cmd := r.ReadCmd(); cmd != TagBoz {
                    r.Fail(errors.New("expected: boz"))
                }
                s.Bar[i] = new(Boz)
                s.Bar[i].ReadBareFrom(r)
            }
        }

        func (s *Foo) WriteBareTo(w *tlschema.Writer) {
            w.WriteInt(len(s.Bar))
            for i := 0; i < len(s.Bar); i++ {
                w.WriteCmd(TagBoz)
                s.Bar[i].WriteBareTo(w)
            }
        }   

        // Boz represents boz from TL schema
        type Boz struct {
        }

        func (s *Boz) Cmd() uint32 {
            return TagBoz
        }

        func (s *Boz) ReadBareFrom(r *tlschema.Reader) {
        }

        func (s *Boz) WriteBareTo(w *tlschema.Writer) {
        }

        func ReadBoxedObjectFrom(r *tlschema.Reader) tl.Object {
            cmd := r.ReadCmd()
            switch cmd {
            case TagFoo:
                s := new(Foo)
                s.ReadBareFrom(r)
                return s
            case TagBoz:
                s := new(Boz)
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
	GenerateGoCode(sch, Options{PackageName: "foo"})
}

func TestTelegram(t *testing.T) {
	sch := tlschema.MustParse(knownschemas.TelegramSchema)
	GenerateGoCode(sch, Options{PackageName: "foo"})
}
