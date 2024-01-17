package iostreams

import "io"

// Loud is unimpacted by quiet being set on the iostreams.
type Loud interface {
	IOStreams

	LoudErr() io.Writer
}

func (s *system) LoudErr() io.Writer {
	return s.err
}

func (t *Testing) LoudErr() io.Writer {
	return t.Error
}

// UseLoud takes an IOStream and if it implements the Load interfaces, it will
// be used instead of the quiet alternatives.
func UseLoud(io IOStreams) IOStreams {
	lo := &loudWrap{
		IOStreams: io,
	}

	if l, ok := io.(Loud); ok {
		lo.l = l
	}

	return lo

}

type loudWrap struct {
	IOStreams
	l Loud
}

func (l *loudWrap) Err() io.Writer {
	if l.l != nil {
		return l.l.LoudErr()
	}

	return l.Err()
}
