package tlschema

import (
	"fmt"
)

// "fmt"
// 	"log"
// 	"strconv"
// 	"strings"

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
	err := sch.Parse(text, ParseOptions{
		FixZeroTags: true,
	})
	if err != nil {
		panic(err)
	}
	return sch
}

type Priority int

const (
	PriorityNormal       Priority = 0
	PriorityJustADefault          = -10
	PriorityOverride              = 10
)

type ParseOptions struct {
	Origin   string
	Priority Priority

	FixZeroTags   bool
	FixZeroTagsIn map[string]bool
	AllowZeroTags bool

	MarkInternal bool

	Alterations *Alterations
}

func (sch *Schema) Parse(text string, options ParseOptions) error {
	defs, err := Parse(text)
	if err != nil {
		return err
	}

	if sch.tagsToCombs == nil {
		sch.tagsToCombs = make(map[uint32]*Comb)
		sch.namesToCombs = make(map[string]*Comb)
		sch.namesToTypes = make(map[string]*Type)
	}

	for _, def := range defs {
		if def.Tag == 0 {
			if options.FixZeroTags || (options.FixZeroTagsIn != nil && options.FixZeroTagsIn[def.CombName.Full()]) {
				err = def.FixTag()
				if err != nil {
					return err
				}
			} else if !options.AllowZeroTags {
				return fmt.Errorf("zero tag in %q from %s", def.OriginalStr, options.Origin)
			}
		}

		if options.Alterations != nil {
			def.Alter(options.Alterations)
		}

		err = def.Simplify()
		if err != nil {
			return err
		}

		if options.MarkInternal {
			def.IsInternal = true
		}

		err = sch.addComb(&Comb{Def: def, Origin: options.Origin, Priority: options.Priority})
		if err != nil {
			return err
		}
	}

	return nil
}

func (sch *Schema) addComb(comb *Comb) error {
	name := comb.CombName.Full()
	if name != "" {
		if alt := sch.namesToCombs[name]; alt != nil {
			if alt.Priority == comb.Priority {
				return fmt.Errorf("duplicate combinators: %s#%08x in %s and %s#%08x in %s", name, alt.Tag, alt.Origin, name, comb.Tag, comb.Origin)
			} else if alt.Priority > comb.Priority {
				return nil
			} else {
				sch.removeComb(alt)
			}
		}
	}

	if comb.Tag != 0 {
		if alt := sch.tagsToCombs[comb.Tag]; alt != nil {
			if alt.Priority == comb.Priority {
				return fmt.Errorf("tag %08x conflict between %s (in %s) and %s (in %s)", comb.Tag, alt.CombName.String(), alt.Origin, comb.CombName.String(), comb.Origin)
			} else if alt.Priority > comb.Priority {
				return nil
			} else {
				sch.removeComb(alt)
			}
		}
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
				return fmt.Errorf("bare result type: %v", comb.ResultType.String())
			}
			typ = &Type{Name: comb.ResultType.Name, Origin: comb.Origin}
			sch.types = append(sch.types, typ)
			sch.namesToTypes[comb.TypeStr] = typ
		}
		typ.Ctors = append(typ.Ctors, comb)
	}

	return nil
}

func (sch *Schema) removeComb(comb *Comb) {
	name := comb.CombName.Full()

	if comb.Tag != 0 {
		delete(sch.tagsToCombs, comb.Tag)
	}

	if name != "" {
		delete(sch.namesToCombs, name)
	}

	for i, c := range sch.combs {
		if c == comb {
			sch.combs = append(sch.combs[:i], sch.combs[i+1:]...)
			break
		}
	}

	if !comb.IsFunc && comb.ResultType.IsJustTypeName() {
		typ := sch.namesToTypes[comb.TypeStr]
		if typ != nil {
			delete(sch.namesToTypes, comb.TypeStr)

			for i, t := range sch.types {
				if t == typ {
					sch.types = append(sch.types[:i], sch.types[i+1:]...)
					break
				}
			}
		}
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
