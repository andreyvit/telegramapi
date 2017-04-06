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
        )

        const (
            TagNearestDc        uint32 = 0x8e1a1775
            TagHelpGetNearestDc        = 0x1fb33026
        )

        type NearestDc struct {
            Country   string
            ThisDc    int
            NearestDc int
        }

        func (s *NearestDc) Cmd() uint32 {
            return TagNearestDc
        }

        func (s *NearestDc) ReadFrom(r *tlschema.Reader) {
            s.Country = r.ReadString()
            s.ThisDc = r.ReadInt()
            s.NearestDc = r.ReadInt()
        }

        func (s *NearestDc) WriteTo(w *tlschema.Writer) {
            w.WriteCmd(TagNearestDc)
            w.WriteString(s.Country)
            w.WriteInt(s.ThisDc)
            w.WriteInt(s.NearestDc)
        }

        type HelpGetNearestDc struct {
        }

        func (s *HelpGetNearestDc) Cmd() uint32 {
            return TagHelpGetNearestDc
        }

        func (s *HelpGetNearestDc) ReadFrom(r *tlschema.Reader) {
        }

        func (s *HelpGetNearestDc) WriteTo(w *tlschema.Writer) {
            w.WriteCmd(TagHelpGetNearestDc)
        }

        func ReadFrom(r *tlschema.Reader) tlschema.Struct {
            switch r.Cmd() {
            case TagNearestDc:
                s := new(NearestDc)
                s.ReadFrom(r)
                return s
            case TagHelpGetNearestDc:
                s := new(HelpGetNearestDc)
                s.ReadFrom(r)
                return s
            default:
                return nil
            }
        }    `
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
