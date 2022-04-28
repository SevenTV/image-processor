package container

import (
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"
)

var TypeAvif = types.NewType("avif", "image/avif")

func init() {
	filetype.AddMatcher(TypeAvif, func(data []byte) bool {
		if len(data) < 12 {
			return false
		}

		return data[0] == 0x00 &&
			data[1] == 0x00 &&
			data[4] == 'f' &&
			data[5] == 't' &&
			data[6] == 'y' &&
			data[7] == 'p' &&
			data[8] == 'a' &&
			data[9] == 'v' &&
			data[10] == 'i' &&
			(data[11] == 's' || data[11] == 'f' || data[11] == 'o')
	})
}

func Match(data []byte) types.Type {
	t, _ := filetype.Match(data)

	return t
}
