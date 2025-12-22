package progress

import (
	"errors"
	"fmt"
	"io"
)

type (
	// Seeker implements a io.ReadSeeker while collecting the
	// read-position inside the contained io.ReadSeeker
	Seeker struct {
		Size     int64
		Progress int64

		next io.ReadSeeker
	}
)

// New creates a new Seeker from the given io.ReadSeeker
func New(next io.ReadSeeker) (*Seeker, error) {
	lastByte, err := next.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("seeking end of reader: %w", err)
	}

	if _, err := next.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seeking start of reader: %w", err)
	}

	return &Seeker{
		next: next,
		Size: lastByte,
	}, nil
}

// Read implements io.ReadSeeker
func (p *Seeker) Read(o []byte) (n int, err error) {
	i, err := p.next.Read(o)
	if err != nil {
		if errors.Is(err, io.EOF) {
			return i, io.EOF
		}
		return i, fmt.Errorf("reading next reader: %w", err)
	}

	p.Progress += int64(i)

	return i, nil
}

// Seek implements io.ReadSeeker
func (p *Seeker) Seek(offset int64, whence int) (int64, error) {
	pos, err := p.next.Seek(offset, whence)
	if err != nil {
		return pos, fmt.Errorf("seeking next reader: %w", err)
	}

	p.Progress = pos
	return pos, nil
}
