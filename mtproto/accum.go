package mtproto

type Accum struct {
	err error
}

func (a *Accum) Error() error {
	return a.err
}

func (a *Accum) Failed() bool {
	return a.err != nil
}

func (a *Accum) Push(err error) {
	if err != nil {
		if a.err == nil {
			a.err = err
		}
	}
}
