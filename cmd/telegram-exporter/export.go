package main

import (
	"bytes"
	"regexp"
	"strings"
	"time"

	"github.com/andreyvit/telegramapi"
)

type Format int

const (
	FormatFavorites Format = iota
)

var dividerRegexp = regexp.MustCompile(`^((\*\s*){2,}|={2,}|-{2,})$`)
var dividerRegexpStart = regexp.MustCompile(`^((\*\s*){2,}|={2,}|-{2,})\s`)

type Exporter struct {
	UserNameAliases map[string]string
	Format          Format
	TimeZone        *time.Location
}

type exportState struct {
	lastDate Date
	hadMsgs  bool
}

func (exp *Exporter) Export(chat *telegramapi.Chat) string {
	var buf bytes.Buffer
	var state exportState

	for _, msg := range chat.Messages.Messages {
		exp.exportMsg(&buf, &state, msg)
	}

	return buf.String()
}

func (exp *Exporter) exportMsg(w *bytes.Buffer, state *exportState, msg *telegramapi.Message) {
	date := MakeDate(msg.Date.Add(-4 * time.Hour).In(exp.TimeZone).Date())
	if state.lastDate.IsZero() || !date.Equal(state.lastDate) {
		if !state.lastDate.IsZero() {
			w.WriteString("\n\n")
		}
		state.lastDate = date
		w.WriteString("###  ")
		w.WriteString(date.String())
		w.WriteString("  ###\n\n")
		state.hadMsgs = false
	}

	text := msg.Text
	text = strings.Replace(text, "\n", "\n    ", -1)
	isMultiline := strings.Contains(text, "\n")

	if dividerRegexp.MatchString(text) {
		if state.hadMsgs {
			w.WriteString("\n")
			w.WriteString("* * *\n")
			w.WriteString("\n")
			state.hadMsgs = false
		}
		return
	} else if dividerRegexpStart.MatchString(text) {
		comment := strings.TrimSpace(dividerRegexpStart.ReplaceAllString(text, ""))
		w.WriteString("\n")
		w.WriteString("* * * " + comment + "\n")
		w.WriteString("\n")
		state.hadMsgs = false
		return
	}

	from, tm := msg.From, msg.Date
	if msg.FwdFrom != nil {
		from = msg.FwdFrom
		tm = msg.FwdDate
	}

	if from != nil {
		name := from.Name()
		if alias := exp.UserNameAliases[name]; alias != "" {
			name = alias
		}
		w.WriteString(name)

		if false {
			w.WriteString(" (")
			w.WriteString(tm.Format("2006-01-02 15:04:05"))
			w.WriteString(")")
		}
		w.WriteString(": ")
	}

	w.WriteString(text)
	w.WriteString("\n")
	if isMultiline {
		w.WriteString("\n")
	}

	state.hadMsgs = true
}
