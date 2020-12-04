// SPDX-FileCopyrightText: © 2020 Etebase Authors
// SPDX-License-Identifier: BSD-3-Clause

package codec

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testObj struct {
	A string
	B int
	C []bool
}

func TestCodec(t *testing.T) {
	var obj = testObj{
		"test", 99, []bool{true, false},
	}

	bz, err := Marshal(obj)
	require.NoError(t, err)

	var got testObj
	require.NoError(t,
		NewDecoder(bytes.NewBuffer(bz)).Decode(&got),
	)

	assert.Equal(t, obj, got)
}
