// SPDX-FileCopyrightText: © 2020 Etebase Authors
// SPDX-License-Identifier: BSD-3-Clause

package codec

import (
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

type Decoder interface {
	Decode(interface{}) error
}

func NewDecoder(r io.Reader) Decoder {
	return msgpack.NewDecoder(r)
}

func Marshal(v interface{}) ([]byte, error) {
	return msgpack.Marshal(v)
}
