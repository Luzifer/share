package main

import (
	"io"
	"io/ioutil"
)

type progressSeeker struct {
	Size     int64
	Progress int64

	o io.ReadSeeker
}

func newProgressSeeker(o io.ReadSeeker) (*progressSeeker, error) {
	if _, err := o.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(o)
	if err != nil {
		return nil, err
	}

	if _, err := o.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}

	return &progressSeeker{
		o:    o,
		Size: int64(len(data)),
	}, nil
}

func (p *progressSeeker) Read(o []byte) (n int, err error) {
	i, err := p.o.Read(o)

	p.Progress += int64(i)

	return i, err
}

func (p *progressSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		p.Progress = offset
	case io.SeekCurrent:
		p.Progress = p.Progress + offset
	case io.SeekEnd:
		p.Progress = p.Size + offset
	}

	return p.o.Seek(offset, whence)
}
