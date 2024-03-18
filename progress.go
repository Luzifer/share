package main

import (
	"errors"
	"fmt"
	"io"
)

type progressSeeker struct {
	Size     int64
	Progress int64

	next io.ReadSeeker
}

func newProgressSeeker(next io.ReadSeeker) (*progressSeeker, error) {
	lastByte, err := next.Seek(0, io.SeekEnd)
	if err != nil {
		return nil, fmt.Errorf("seeking end of reader: %w", err)
	}

	if _, err := next.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seeking start of reader: %w", err)
	}

	return &progressSeeker{
		next: next,
		Size: lastByte,
	}, nil
}

func (p *progressSeeker) Read(o []byte) (n int, err error) {
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

func (p *progressSeeker) Seek(offset int64, whence int) (int64, error) {
	pos, err := p.next.Seek(offset, whence)
	if err != nil {
		return pos, fmt.Errorf("seeking next reader: %w", err)
	}

	p.Progress = pos
	return pos, nil
}
