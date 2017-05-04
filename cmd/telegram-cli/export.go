package main

import (
	"bytes"
	"strings"
	"time"

	"github.com/andreyvit/telegramapi"
)

type Format int

const (
	FormatFavorites Format = iota
)

type Exporter struct {
	UserNameAliases map[string]string
	Format          Format
	TimeZone        *time.Location

	lastDate Date
}

func (exp *Exporter) Export(chat *telegramapi.Chat) string {
	var buf bytes.Buffer

	for _, msg := range chat.Messages.Messages {
		exp.exportMsg(&buf, msg)
	}

	return buf.String()
}

func (exp *Exporter) exportMsg(w *bytes.Buffer, msg *telegramapi.Message) {
	date := MakeDate(msg.Date.Add(-4 * time.Hour).In(exp.TimeZone).Date())
	if exp.lastDate.IsZero() || !date.Equal(exp.lastDate) {
		if !exp.lastDate.IsZero() {
			w.WriteString("\n\n")
		}
		exp.lastDate = date
		w.WriteString("###  ")
		w.WriteString(date.String())
		w.WriteString("  ###\n\n")
	}

	text := msg.Text
	text = strings.Replace(text, "\n", "\n    ", -1)
	isMultiline := strings.Contains(text, "\n")

	from, tm := msg.From, msg.Date
	if msg.FwdFrom != nil {
		from = msg.FwdFrom
		tm = msg.FwdDate
	}

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
	w.WriteString(text)
	w.WriteString("\n")
	if isMultiline {
		w.WriteString("\n")
	}
}
