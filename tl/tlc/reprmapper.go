package tlc

import (
	"bytes"

	"github.com/andreyvit/telegramapi/tl/tlschema"
)

type finalizationState int

const (
	open finalizationState = iota
	finalizing
	finalized
)

type ReprMapper struct {
	prefix string

	schema    *tlschema.Schema
	typeReprs map[string]GenericRepr

	typeOverrides map[string]string

	contribByName map[string]Contributor
	contributors  []Contributor
	finalized     map[string]bool

	finState finalizationState
}

func NewReprMapper(sch *tlschema.Schema) *ReprMapper {
	rm := &ReprMapper{
		prefix:    "TL",
		schema:    sch,
		typeReprs: make(map[string]GenericRepr),
		typeOverrides: map[string]string{
			"resPQ:pq":                      "bigint_",
			"p_q_inner_data:pq":             "bigint_",
			"p_q_inner_data:p":              "bigint_",
			"p_q_inner_data:q":              "bigint_",
			"server_DH_inner_data:dh_prime": "bigint_",
			"server_DH_inner_data:g_a":      "bigint_",
			"client_DH_inner_data:g_b":      "bigint_",

			"server_DH_inner_data:server_time": "unixtime_",
		},
		contribByName: make(map[string]Contributor),
		finalized:     make(map[string]bool),
	}

	rm.schema.MustParse("string ? = String")
	rm.schema.MustParse("int ? = Int")
	rm.schema.MustParse("long ? = Long")
	rm.schema.MustParse("int128 ? = Int128")
	rm.schema.MustParse("int256 ? = Int256")
	rm.schema.MustParse("bytes ? = Bytes")
	rm.schema.MustParse("bigint_ ? = BigInt_")
	rm.schema.MustParse("unixtime_ ? = UnixTime_")
	rm.schema.MustParse("object ? = Object")
	rm.schema.MustParse("vector#1cb5c415 ? = Vector")

	rm.typeReprs["String"] = &StringRepr{}
	rm.typeReprs["Int"] = &IntRepr{}
	rm.typeReprs["Long"] = &LongRepr{}
	rm.typeReprs["Int128"] = &Int128Repr{}
	rm.typeReprs["Int256"] = &Int256Repr{}
	rm.typeReprs["Bytes"] = &BytesRepr{}
	rm.typeReprs["BigInt_"] = &BigIntRepr{}
	rm.typeReprs["UnixTime_"] = &UnixTimeRepr{}
	rm.typeReprs["Object"] = &ObjectRepr{}
	rm.typeReprs["Vector"] = &GenericVectorRepr{}

	rm.analyze()
	return rm
}

func (rm *ReprMapper) analyze() {
	for _, typ := range rm.schema.Types() {
		rm.AddType(typ)
	}

	for _, comb := range rm.schema.Funcs() {
		sr := &StructRepr{
			TLName: comb.CombName.Full(),
			GoName: rm.prefix + comb.CombName.GoName(),
			Ctor:   comb,
		}
		rm.AddContributor(sr)
	}

	rm.Finalize()
}

func (rm *ReprMapper) AddContributor(c Contributor) Contributor {
	if rm.finState == finalized {
		panic("cannot add new types after Finalize()")
	}

	id := c.InternalTypeID()
	if prev := rm.contribByName[id]; prev != nil {
		return prev
	}
	rm.contribByName[id] = c
	rm.contributors = append(rm.contributors, c)
	return c
}

func (rm *ReprMapper) finalizeContributor(c Contributor) {
	id := c.InternalTypeID()
	if rm.finalized[id] {
		return
	}
	rm.finalized[id] = true

	if rm.finState == finalized {
		panic("cannot finalize new contributors after Finalize() returns")
	}

	c.Resolve(rm)
}

func (rm *ReprMapper) Finalize() {
	if rm.finState != open {
		return
	}
	rm.finState = finalizing

	// cannot use range here because new contributors may be added during this loop
	for i := 0; i < len(rm.contributors); i++ {
		rm.finalizeContributor(rm.contributors[i])
	}

	rm.finState = finalized
}

func (rm *ReprMapper) Specialize(gr GenericRepr, typeExpr tlschema.TypeExpr) Repr {
	rm.finalizeContributor(gr)
	repr := gr.Specialize(typeExpr)
	if repr != nil {
		repr = rm.AddContributor(repr).(Repr)
	}
	return repr
}

func (rm *ReprMapper) ResolveTypeExpr(expr tlschema.TypeExpr, context string) Repr {
	if override := rm.typeOverrides[context]; override != "" {
		expr.Name = tlschema.MakeScopedName(override)
	}

	if expr.Name.IsBare() {
		comb := rm.schema.ByName(expr.Name.Full())
		if comb == nil {
			repr := &UnsupportedRepr{expr.String(), "implied constructor not found for bare type"}
			return rm.AddContributor(repr).(Repr)
		}

		expr.IsPercent = true
		expr.Name = comb.ResultType.Name
	}

	gr := rm.typeReprs[expr.Name.Full()]
	if gr == nil {
		repr := &UnsupportedRepr{expr.String(), "unknown type"}
		return rm.AddContributor(repr).(Repr)
	}

	repr := rm.Specialize(gr, expr)
	if repr != nil {
		return repr
	}

	repr = &UnsupportedRepr{expr.String(), "failed to specialize"}
	return rm.AddContributor(repr).(Repr)
}

func (rm *ReprMapper) FindType(name string) *tlschema.Type {
	return rm.schema.Type(name)
}

func (rm *ReprMapper) FindComb(name string) *tlschema.Comb {
	return rm.schema.ByName(name)
}

func (rm *ReprMapper) AddType(typ *tlschema.Type) {
	if rm.typeReprs[typ.Name.Full()] != nil {
		return
	}

	repr := rm.pickTypeRepr(typ)
	if repr == nil {
		return
	}

	rm.typeReprs[typ.Name.Full()] = repr
	rm.AddContributor(repr)
}

func (rm *ReprMapper) AllTypeIDs() []string {
	if rm.finState != finalized {
		panic("Finalize() must be called before AllTypeIDs()")
	}

	var result []string
	for _, c := range rm.contributors {
		result = append(result, c.InternalTypeID())
	}
	return result
}

func (rm *ReprMapper) GoImports() []string {
	if rm.finState != finalized {
		panic("Finalize() must be called before GoImports()")
	}

	var result []string
	for _, c := range rm.contributors {
		result = append(result, c.GoImports()...)
	}
	return result
}

func (rm *ReprMapper) AppendGoDefs(buf *bytes.Buffer, options CodeGenOptions) {
	if rm.finState != finalized {
		panic("Finalize() must be called before AppendGoDefs()")
	}

	for _, c := range rm.contributors {
		c.AppendGoDefs(buf, options)
	}

	buf.WriteString("\n")
	buf.WriteString("func ReadBoxedObjectFrom(r *tl.Reader) tl.Object {\n")
	buf.WriteString("\tcmd := r.ReadCmd()\n")
	buf.WriteString("\tswitch cmd {\n")
	for _, c := range rm.contributors {
		if gr, ok := c.(GenericRepr); ok {
			gr.AppendSwitchCase(buf, "\t")
		}
	}
	buf.WriteString("\tdefault:\n")
	buf.WriteString("\t\treturn nil\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")

	if !options.SkipUtils {
		buf.WriteString("\n")
		buf.WriteString("func ReadLimitedBoxedObjectFrom(r *tl.Reader, cmds ...uint32) tl.Object {\n")
		buf.WriteString("\tif r.ExpectCmd(cmds...) {\n")
		buf.WriteString("\t\treturn ReadBoxedObjectFrom(r)\n")
		buf.WriteString("\t} else {\n")
		buf.WriteString("\t\treturn nil\n")
		buf.WriteString("\t}\n")
		buf.WriteString("}\n")
	}
}

func (rm *ReprMapper) pickTypeRepr(typ *tlschema.Type) GenericRepr {
	switch len(typ.Ctors) {
	case 0:
		panic("type has no constructors")
	case 1:
		ctor := typ.Ctors[0]
		if ctor.IsWeird {
			return nil
		}
		return &StructRepr{
			TLName: ctor.CombName.Full(),
			GoName: rm.prefix + ctor.CombName.GoName(),
			Ctor:   ctor,
		}
	default:
		funcname := "Is" + rm.prefix + typ.Name.GoName()

		var structs []*StructRepr
		for _, ctor := range typ.Ctors {
			structs = append(structs, &StructRepr{
				TLName:           ctor.CombName.Full(),
				GoName:           rm.prefix + ctor.CombName.GoName(),
				Ctor:             ctor,
				GoMarkerFuncName: funcname,
			})
		}

		return &MultiCtorRepr{
			TLName:           typ.Name.Full(),
			GoName:           rm.prefix + typ.Name.GoName() + "Type",
			Structs:          structs,
			GoMarkerFuncName: funcname,
		}
	}
}
