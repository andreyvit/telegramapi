package telegramapi

import (
	"github.com/andreyvit/telegramapi/mtproto"
)

func processDCs(config *mtproto.TLConfig) []*DC {
	var dcs []*DC
	found := make(map[int]bool)

	for _, opt := range config.DCOptions {
		if opt.IPv6() {
			continue
		}
		if opt.MediaOnly() {
			continue
		}
		if found[opt.ID] {
			continue
		}

		found[opt.ID] = true
		dcs = append(dcs, &DC{
			ID: opt.ID,
			PrimaryAddr: Addr{
				IP:   opt.IPAddress,
				Port: opt.Port,
			},
		})
	}

	return dcs
}
