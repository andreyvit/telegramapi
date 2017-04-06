package tlc

import (
	"bytes"
	"log"

	"github.com/andreyvit/telegramapi/tl/tlschema"
)

type ReprMapper struct {
	schema    *tlschema.Schema
	typeReprs map[string]GenericRepr
	funcReprs map[uint32]*StructRepr
	reprs     []GenericRepr

	typeOverrides map[string]string
}

func NewReprMapper(sch *tlschema.Schema) *ReprMapper {
	rm := &ReprMapper{
		schema:    sch,
		typeReprs: make(map[string]GenericRepr),
		funcReprs: make(map[uint32]*StructRepr),
		typeOverrides: map[string]string{
			"ResPQ:pq":                      "bigint_",
			"P_Q_inner_data:pq":             "bigint_",
			"P_Q_inner_data:p":              "bigint_",
			"P_Q_inner_data:q":              "bigint_",
			"Server_DH_inner_data:dh_prime": "bigint_",
			"Server_DH_inner_data:g_a":      "bigint_",
			"Client_DH_Inner_Data:g_b":      "bigint_",

			"Server_DH_inner_data:server_time": "unixtime_",
		},
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
		if !comb.ResultType.IsJustTypeName() {
			log.Print("cannot map result of " + comb.CombName.Full() + ": type " + comb.ResultType.String())
		}

		sr := &StructRepr{
			TLName: comb.CombName.Full(),
			GoName: comb.CombName.GoName(),
			Ctor:   comb,
		}
		rm.addRepr(sr)
	}

	for _, repr := range rm.reprs {
		repr.Resolve(rm)
	}
}

func (rm *ReprMapper) findGenericRepr(name string, context string) GenericRepr {
	if override := rm.typeOverrides[context]; override != "" {
		name = override
	}

	return rm.typeReprs[name]
}

func (rm *ReprMapper) ResolveTypeExpr(expr tlschema.TypeExpr, context string) Repr {
	if expr.Name.IsBare() {
		comb := rm.schema.ByName(expr.Name.Full())
		if comb == nil {
			return &UnsupportedRepr{expr.String(), "implied constructor not found for bare type"}
		}

		expr.IsPercent = true
		expr.Name = comb.ResultType.Name
	}

	gr := rm.findGenericRepr(expr.Name.Full(), context)
	if gr == nil {
		return &UnsupportedRepr{expr.String(), "unknown type"}
	}

	gr.Resolve(rm)

	repr := gr.Specialize(expr)
	if repr != nil {
		return repr
	}

	return &UnsupportedRepr{expr.String(), "failed to specialize"}
}

func (rm *ReprMapper) FindType(name string) *tlschema.Type {
	return rm.schema.Type(name)
}

func (rm *ReprMapper) FindComb(name string) *tlschema.Comb {
	return rm.schema.ByName(name)
}

func (rm *ReprMapper) AddType(typ *tlschema.Type) {
	repr := rm.pick(typ)
	if repr == nil {
		return
	}
	rm.typeReprs[typ.Name.Full()] = repr
	rm.addRepr(repr)
}

func (rm *ReprMapper) GoImports() []string {
	var result []string
	for _, repr := range rm.reprs {
		result = append(result, repr.GoImports()...)
	}
	return result
}

func (rm *ReprMapper) AppendGoDefs(buf *bytes.Buffer) {
	for _, repr := range rm.reprs {
		repr.AppendGoDefs(buf)
	}

	buf.WriteString("\n")
	buf.WriteString("func ReadObjectFrom(r *tlschema.Reader) tl.Object {\n")
	buf.WriteString("\tswitch r.Cmd() {\n")
	for _, repr := range rm.reprs {
		repr.AppendSwitchCase(buf, "\t")
	}
	buf.WriteString("\tdefault:\n")
	buf.WriteString("\t\treturn nil\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")
}

func (rm *ReprMapper) addRepr(repr GenericRepr) {
	rm.reprs = append(rm.reprs, repr)
}

func (rm *ReprMapper) pick(typ *tlschema.Type) GenericRepr {
	if gr := rm.typeReprs[typ.Name.Full()]; gr != nil {
		return gr
	}

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
			GoName: typ.Name.GoName(),
			Ctor:   ctor,
		}
	}

	return &UnsupportedRepr{typ.String(), "multiple ctors not implemented yet"}
}
