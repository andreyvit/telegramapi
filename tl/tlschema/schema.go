package tlschema

import (
// "fmt"
// 	"log"
// 	"strconv"
// 	"strings"
)

type Schema struct {
	combs        []*Comb
	funcs        []*Comb
	tagsToCombs  map[uint32]*Comb
	namesToCombs map[string]*Comb

	types        []*Type
	namesToTypes map[string]*Type
}

func (sch *Schema) Combs() []*Comb {
	return sch.combs
}

func (sch *Schema) Funcs() []*Comb {
	return sch.funcs
}

func (sch *Schema) ByTag(tag uint32) *Comb {
	return sch.tagsToCombs[tag]
}

func (sch *Schema) ByName(name string) *Comb {
	return sch.namesToCombs[name]
}

func (sch *Schema) Types() []*Type {
	return sch.types
}

func (sch *Schema) Type(name string) *Type {
	return sch.namesToTypes[name]
}

// func RegisterCmd(cmd uint32, fullname, def string) {
// 	if cmdToInfo[cmd] != nil {
// 		panic("duplicate command def")
// 	}
// 	if fullnameToCmd[fullname] != 0 {
// 		panic("duplicate command def (name)")
// 	}
// 	if cmd == 0 {
// 		panic("cmd cannot be zero")
// 	}

// 	cinfo := &cmdInfo{cmd, fullname}
// 	cmds = append(cmds, cinfo)
// 	cmdToInfo[cmd] = cinfo
// 	fullnameToCmd[fullname] = cmd
// }

func MustParse(text string) *Schema {
	sch := new(Schema)
	sch.MustParse(text)
	return sch
}

func (sch *Schema) MustParse(text string) {
	defs, err := Parse(text)
	if err != nil {
		panic(err)
	}

	if sch.tagsToCombs == nil {
		sch.tagsToCombs = make(map[uint32]*Comb)
		sch.namesToCombs = make(map[string]*Comb)
		sch.namesToTypes = make(map[string]*Type)
	}

	for _, def := range defs {
		sch.addComb(&Comb{Def: def})
	}
}

func (sch *Schema) addComb(comb *Comb) {
	if comb.Tag != 0 && sch.tagsToCombs[comb.Tag] != nil {
		return
		// 	panic(fmt.Sprintf("tag %08x conflict between %s and %s", comb.Tag, sch.tagsToCombs[comb.Tag].CombName.String(), comb.CombName.String()))
	}
	name := comb.CombName.Full()
	if name != "" && sch.namesToCombs[name] != nil {
		return
		// panic(fmt.Sprintf("name conflict for %q between %08x and %08x", name, comb.Tag, sch.namesToCombs[name].Tag))
	}

	sch.combs = append(sch.combs, comb)
	if comb.IsFunc {
		sch.funcs = append(sch.funcs, comb)
	}

	if comb.Tag != 0 {
		sch.tagsToCombs[comb.Tag] = comb
	}

	if name != "" {
		sch.namesToCombs[name] = comb
	}

	if !comb.IsFunc && comb.ResultType.IsJustTypeName() {
		comb.TypeStr = comb.ResultType.Name.Full()

		typ := sch.namesToTypes[comb.TypeStr]
		if typ == nil {
			if comb.ResultType.Name.IsBare() {
				panic("bare result type: " + comb.ResultType.String())
			}
			typ = &Type{Name: comb.ResultType.Name}
			sch.types = append(sch.types, typ)
			sch.namesToTypes[comb.TypeStr] = typ
		}
		typ.Ctors = append(typ.Ctors, comb)
	}
}

// func (sch *Schema) AddLine(line string) {
// }

// func Cmd(fullname string) uint32 {
// 	cmd := fullnameToCmd[fullname]
// 	if cmd == 0 {
// 		panic(fmt.Sprintf("command not found: %#v", fullname))
// 	}
// 	return cmd
// }

// func CmdName(cmd uint32) string {
// 	if cinfo := cmdToInfo[cmd]; cinfo != nil {
// 		return cinfo.fullname
// 	} else {
// 		return fmt.Sprintf("#%08x", cmd)
// 	}
// }

// func DescribeCmd(cmd uint32) string {
// 	if cmd == 0 {
// 		return "none"
// 	} else if cinfo := cmdToInfo[cmd]; cinfo != nil {
// 		return fmt.Sprintf("%s#%08x", cinfo.fullname, cmd)
// 	} else {
// 		return fmt.Sprintf("#%08x", cmd)
// 	}
// }

// func DescribeCmdOfPayload(b []byte) string {
// 	return DescribeCmd(CmdOfPayload(b))
// }

// func init() {
// 	AddSchema(mtprotoSchema)
// 	AddSchema(apiSchema)
// }
