package telegramapi

import (
	"github.com/andreyvit/telegramapi/mtproto"
)

func updateDCs(dcs map[int]*DCState, config *mtproto.TLConfig) {
	oldIDs := make(map[int]bool)
	for id, _ := range dcs {
		oldIDs[id] = true
	}

	for _, opt := range config.DCOptions {
		if opt.IPv6() {
			continue
		}
		if opt.MediaOnly() {
			continue
		}

		dc := dcs[opt.ID]
		if dc == nil {
			dc = &DCState{ID: opt.ID}
			dcs[dc.ID] = dc
		}

		dc.PrimaryAddr = Addr{
			IP:   opt.IPAddress,
			Port: opt.Port,
		}

		delete(oldIDs, dc.ID)
	}

	for id, _ := range oldIDs {
		delete(dcs, id)
	}
}
