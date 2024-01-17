// package promptio wraps a test IOStream to make it safe to use the input with
// promptui.
package promptio

import (
	"io"

	"github.com/hashicorp/hcp/internal/pkg/iostreams"
)

// PromptIO wraps a testing iostream to make it safe to test promptui.
type PromptIO struct {
	*iostreams.Testing
}

// Wrap wraps the testing iostreams and returns an iostream safe for use with
// promptui.
func Wrap(t *iostreams.Testing) *PromptIO {
	return &PromptIO{t}
}

func (p *PromptIO) In() io.Reader {
	return &stdinBuffer{p.Testing.In()}
}

// stdinBuffer is a buffer that forces a read of one byte at a time.
type stdinBuffer struct {
	io.Reader
}

func (b *stdinBuffer) Close() error {
	return nil
}

func (b *stdinBuffer) Read(dst []byte) (int, error) {
	buf := make([]byte, 1)
	n, err := b.Reader.Read(buf)
	if err != nil {
		return n, err
	}

	copy(dst, buf)
	return n, nil
}
