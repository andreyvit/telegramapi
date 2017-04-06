package tlc

import (
	"github.com/andreyvit/telegramapi/tl/tlschema"
)

func specializeBare(repr Repr, comb *tlschema.Comb, typ tlschema.TypeExpr) Repr {
	if typ.IsBare() {
		return repr
	} else {
		return &BoxedRepr{Comb: comb, ItemRepr: repr}
	}
}

func specializeOnlyBare(repr Repr, typ tlschema.TypeExpr) Repr {
	if typ.IsBare() {
		return repr
	} else {
		return &UnsupportedRepr{typ.String(), "only bare type supported"}
	}
}

func specializeOnlyNonBare(repr Repr, typ tlschema.TypeExpr) Repr {
	if !typ.IsBare() {
		return repr
	} else {
		return &UnsupportedRepr{typ.String(), "bare type not supported"}
	}
}
