package main

import (
	"bytes"
	"strings"

	"github.com/andreyvit/telegramapi"
)

type Exporter struct {
	UserNameAliases map[string]string
}

func (exp *Exporter) Export(chat *telegramapi.Chat) string {
	var buf bytes.Buffer

	for _, msg := range chat.Messages.Messages {
		exp.exportMsg(&buf, msg)
	}

	return buf.String()
}

func (exp *Exporter) exportMsg(w *bytes.Buffer, msg *telegramapi.Message) {
	text := msg.Text
	text = strings.Replace(text, "\n", "\n    ", -1)
	isMultiline := strings.Contains(text, "\n")

	from, date := msg.From, msg.Date
	if msg.FwdFrom != nil {
		from = msg.FwdFrom
		date = msg.FwdDate
	}

	w.WriteString(from.Name())
	w.WriteString(" (")
	w.WriteString(date.Format("2006-01-02 15:04:05"))
	w.WriteString(")")
	w.WriteString(": ")
	w.WriteString(text)
	w.WriteString("\n")
	if isMultiline {
		w.WriteString("\n")
	}
}
