package tlc

import (
	"bytes"
	"log"

	"github.com/andreyvit/telegramapi/tl/tlschema"
)

type ReprMapper struct {
	schema    *tlschema.Schema
	typeReprs map[string]Repr
	funcReprs map[uint32]*StructRepr
	reprs     []Repr

	typeOverrides map[string]string
}

func NewReprMapper(sch *tlschema.Schema) *ReprMapper {
	rm := &ReprMapper{
		schema:    sch,
		typeReprs: make(map[string]Repr),
		funcReprs: make(map[uint32]*StructRepr),
		typeOverrides: map[string]string{
			"ResPQ:pq":                      "bigint",
			"P_Q_inner_data:pq":             "bigint",
			"P_Q_inner_data:p":              "bigint",
			"P_Q_inner_data:q":              "bigint",
			"Server_DH_inner_data:dh_prime": "bigint",
			"Server_DH_inner_data:g_a":      "bigint",
			"Client_DH_Inner_Data:g_b":      "bigint",
		},
	}
	rm.analyze()
	return rm
}

func (rm *ReprMapper) analyze() {
	for _, typ := range rm.schema.Types() {
		repr := rm.pick(typ)
		rm.typeReprs[typ.Name.Full()] = repr
		rm.addRepr(repr)
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

func (rm *ReprMapper) TryResolveTypeName(name string, context string) Repr {
	if override := rm.typeOverrides[context]; override != "" {
		name = override
	}

	switch name {
	case "string":
		return &StringRepr{}
	case "int":
		return &IntRepr{}
	case "long":
		return &LongRepr{}
	case "int128":
		return &Int128Repr{}
	case "int256":
		return &Int256Repr{}
	case "bytes":
		return &BytesRepr{}
	case "bigint": // pseudo-type from typeOverrides
		return &BigIntRepr{}
	case "Object":
		return &ObjectRepr{}
	}

	return rm.typeReprs[name]
}

func (rm *ReprMapper) ResolveTypeExpr(expr tlschema.TypeExpr, context string) Repr {
	if expr.IsJustTypeName() {
		name := expr.Name.Full()
		repr := rm.TryResolveTypeName(name, context)
		if repr != nil {
			return repr
		} else {
			return &UnknownTypeRefRepr{name}
		}
	} else {
		return &UndefinedRepr{}
	}
}

func (rm *ReprMapper) AddType(typ *tlschema.Type) {
	repr := rm.pick(typ)
	rm.typeReprs[typ.Name.Full()] = repr
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
	buf.WriteString("func ReadFrom(r *tlschema.Reader) tlschema.Struct {\n")
	buf.WriteString("\tswitch r.Cmd() {\n")
	for _, repr := range rm.reprs {
		repr.AppendSwitchCase(buf, "\t", "r")
	}
	buf.WriteString("\tdefault:\n")
	buf.WriteString("\t\treturn nil\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")
}

func (rm *ReprMapper) addRepr(repr Repr) {
	rm.reprs = append(rm.reprs, repr)
}

func (rm *ReprMapper) pick(typ *tlschema.Type) Repr {
	if repr := rm.TryResolveTypeName(typ.Name.Full(), ""); repr != nil {
		return repr
	}

	switch len(typ.Ctors) {
	case 1:
		return &StructRepr{
			TLName: typ.Name.Full(),
			GoName: typ.Name.GoName(),
			Ctor:   typ.Ctors[0],
		}
	}

	return &UndefinedRepr{}
}
