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
	typeAliases   map[string]string

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
			"req_DH_params:p":               "bigint_",
			"req_DH_params:q":               "bigint_",
			"server_DH_inner_data:dh_prime": "bigint_",
			"server_DH_inner_data:g_a":      "bigint_",
			"client_DH_inner_data:g_b":      "bigint_",

			"server_DH_inner_data:server_time": "unixtime_",
		},
		typeAliases: map[string]string{
			"#": "nat",
		},
		contribByName: make(map[string]Contributor),
		finalized:     make(map[string]bool),
	}

	rm.AddSpecialType("True", "true#3fedd339 = True", &TrueRepr{}, false)
	rm.AddSpecialType("Bool", "boolFalse#bc799737 = Bool;\nboolTrue#997275b5 = Bool;", &BoolRepr{}, false)
	rm.AddSpecialType("String", "string ? = String", &StringRepr{}, false)
	rm.AddSpecialType("Nat", "nat ? = Nat", &NatRepr{}, true)
	rm.AddSpecialType("Int", "int ? = Int", &IntRepr{}, false)
	rm.AddSpecialType("Long", "long ? = Long", &LongRepr{}, false)
	rm.AddSpecialType("Int128", "int128 ? = Int128", &Int128Repr{}, true)
	rm.AddSpecialType("Int256", "int256 ? = Int256", &Int256Repr{}, true)
	rm.AddSpecialType("Double", "double ? = Double", &DoubleRepr{}, false)
	rm.AddSpecialType("Bytes", "bytes ? = Bytes", &BytesRepr{}, false)
	rm.AddSpecialType("BigInt_", "bigint_ ? = BigInt_", &BigIntRepr{}, true)
	rm.AddSpecialType("UnixTime_", "unixtime_ ? = UnixTime_", &UnixTimeRepr{}, true)
	rm.AddSpecialType("Object", "object ? = Object", &ObjectRepr{}, false)
	rm.AddSpecialType("Vector", "vector#1cb5c415 ? = Vector", &GenericVectorRepr{}, false)

	rm.analyze()
	return rm
}

func (rm *ReprMapper) AddSpecialType(name, def string, gr GenericRepr, internal bool) {
	err := rm.schema.Parse(def, tlschema.ParseOptions{
		Origin:       "builtin",
		FixZeroTags:  true,
		Priority:     tlschema.PriorityOverride,
		MarkInternal: internal,
	})
	if err != nil {
		panic(err)
	}
	rm.typeReprs[name] = gr
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
	if alias := rm.typeAliases[expr.Name.Full()]; alias != "" {
		expr.Name = tlschema.MakeScopedName(alias)
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

	if !options.SkipSwitch {
		buf.WriteString("\n")
		buf.WriteString("var Schema = &tl.Schema{\n")
		buf.WriteString("\tFactory: func(cmd uint32) tl.Object {\n")
		buf.WriteString("\t\tswitch cmd {\n")
		for _, c := range rm.contributors {
			if gr, ok := c.(GenericRepr); ok {
				gr.AppendSwitchCase(buf, "\t\t")
			}
		}
		buf.WriteString("\t\tdefault:\n")
		buf.WriteString("\t\t\treturn nil\n")
		buf.WriteString("\t\t}\n")
		buf.WriteString("\t},\n")
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
