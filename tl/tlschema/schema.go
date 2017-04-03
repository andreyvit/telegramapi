package tlschema

// import (
// 	"fmt"
// 	"log"
// 	"strconv"
// 	"strings"
// )

// type Schema struct {
// }

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

// func (sch *Schema) AddText(schema string) {
// 	for _, line := range strings.Split(schema, "\n") {
// 		line = strings.TrimSpace(line)
// 		if len(line) == 0 {
// 			continue
// 		}
// 		if line[0:2] == "//" {
// 			continue
// 		}
// 		if line[0:3] == "---" {
// 			continue
// 		}

// 		AddSchemaLine(line)
// 	}
// }

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
