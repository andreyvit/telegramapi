package mtproto

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/andreyvit/telegramapi/tl"
)

//go:generate go install ../tl/cmd/tlc
//go:generate tlc -o generated.go mtproto telegram

type cmdInfo struct {
	cmd      uint32
	fullname string
}

var cmds []*cmdInfo
var cmdToInfo = make(map[uint32]*cmdInfo)
var fullnameToCmd = make(map[string]uint32)

func RegisterCmd(cmd uint32, fullname, def string) {
	if cmdToInfo[cmd] != nil {
		panic("duplicate command def")
	}
	if fullnameToCmd[fullname] != 0 {
		panic("duplicate command def (name)")
	}
	if cmd == 0 {
		panic("cmd cannot be zero")
	}

	cinfo := &cmdInfo{cmd, fullname}
	cmds = append(cmds, cinfo)
	cmdToInfo[cmd] = cinfo
	fullnameToCmd[fullname] = cmd
}

func AddSchema(schema string) {
	for _, line := range strings.Split(schema, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		if line[0:2] == "//" {
			continue
		}
		if line[0:3] == "---" {
			continue
		}

		AddSchemaLine(line)
	}
}

func AddSchemaLine(line string) {
	line = strings.TrimSpace(line)

	if strings.HasSuffix(line, ";") {
		line = line[:len(line)-1]
	}

	fields := strings.Fields(line)
	name, cmd := parseCombinatorName(fields[0])
	if cmd != 0 {
		RegisterCmd(cmd, name, line)
	}
}

func Cmd(fullname string) uint32 {
	cmd := fullnameToCmd[fullname]
	if cmd == 0 {
		panic(fmt.Sprintf("command not found: %#v", fullname))
	}
	return cmd
}

func CmdName(cmd uint32) string {
	if cinfo := cmdToInfo[cmd]; cinfo != nil {
		return cinfo.fullname
	} else {
		return fmt.Sprintf("#%08x", cmd)
	}
}

func DescribeCmd(cmd uint32) string {
	if cmd == 0 {
		return "none"
	} else if cinfo := cmdToInfo[cmd]; cinfo != nil {
		return fmt.Sprintf("%s#%08x", cinfo.fullname, cmd)
	} else {
		return fmt.Sprintf("#%08x", cmd)
	}
}

func DescribeCmdOfPayload(b []byte) string {
	return DescribeCmd(tl.CmdOfPayload(b))
}

func parseCombinatorName(s string) (string, uint32) {
	idx := strings.IndexRune(s, '#')
	if idx < 0 {
		return s, 0
	}

	name := s[:idx]
	cmdstr := s[idx+1:]
	if len(cmdstr) > 8 {
		log.Panicf("invalid schema, cmd hex code > 8 chars in %#v", s)
	}
	cmd, err := strconv.ParseUint(cmdstr, 16, 32)
	if err != nil {
		log.Panicf("invalid schema, cannot parse hex in %#v: %v", s, err)
	}
	return name, uint32(cmd)
}

func init() {
	AddSchema(mtprotoSchema)
	AddSchema(apiSchema)
}
