package pages

import (
	"bytes"
)

type ByteBuf struct {
	*bytes.Buffer
}

func NewByteBuf(byte []byte) *ByteBuf {
	return &ByteBuf{
		bytes.NewBuffer(byte),
	}
}

func (b ByteBuf) Close() error {
	return nil
}
