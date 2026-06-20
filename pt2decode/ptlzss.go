package main

func (pt2d *Pt2Decoder) DecodeLZ77(payload []byte, outSize int) []byte {
	if len(payload) < 4 {
		return nil
	}
	ctrl := le32(payload, 0)
	remaining := int(ctrl & 0x00FFFFFF)
	shift := int((ctrl >> 24) & 0x0F)
	mask := (1 << shift) - 1
	extended := (ctrl & 0x40000000) != 0
	sentinel := -1
	if extended {
		sentinel = mask
	}

	out := make([]byte, outSize)
	src := 4
	dst := 0
	flags := 0xFFFF

	for remaining > 0 {
		if flags == 0xFFFF {
			if src+2 > len(payload) {
				return nil
			}
			flags = 0xFFFF0000 | int(le16(payload, src))
			src += 2
		}

		if flags&1 != 0 {
			if src >= len(payload) || dst >= len(out) {
				return nil
			}
			out[dst] = payload[src]
			src++
			dst++
			remaining--
		} else {
			if src+2 > len(payload) {
				return nil
			}
			token := int(le16(payload, src))
			src += 2

			distance := token >> shift
			lenCode := token & mask
			if distance == 0 {
				if src+2 > len(payload) {
					return nil
				}
				distance = int(le16(payload, src))
				src += 2
			}
			if distance <= 0 {
				return nil
			}
			if sentinel >= 0 && lenCode == sentinel {
				if src >= len(payload) {
					return nil
				}
				lenCode = int(payload[src]) + sentinel
				src++
			}

			length := lenCode + 3
			if remaining < length || dst+length > len(out) {
				return nil
			}

			copyFrom := dst - distance
			if copyFrom < 0 {
				return nil
			}
			for i := 0; i < length; i++ {
				out[dst+i] = out[copyFrom+i]
			}
			dst += length
			remaining -= length
		}

		flags >>= 1
	}
	return out
}

func (pt2d *Pt2Decoder) DecodeLzss(payload []byte, outSize int) []byte {
	if len(payload) < 4 {
		return nil
	}
	ctrl := le32(payload, 0)
	remaining := int(ctrl & 0x00FFFFFF)
	shift := int((ctrl >> 24) & 0x0F)
	mask := (1 << shift) - 1
	extended := (ctrl & 0x40000000) != 0
	sentinel := -1
	if extended {
		sentinel = mask
	}

	out := make([]byte, outSize)
	src := 4
	dst := 0
	flags := 0xFFFF

	for remaining > 0 {
		if flags == 0xFFFF {
			if src+2 > len(payload) {
				return nil
			}
			flags = 0xFFFF0000 | int(le16(payload, src))
			src += 2
		}

		if flags&1 != 0 {
			if src >= len(payload) || dst >= len(out) {
				return nil
			}
			out[dst] = payload[src]
			src++
			dst++
			remaining--
		} else {
			if src+2 > len(payload) {
				return nil
			}
			token := int(le16(payload, src))
			src += 2

			distance := token >> shift
			lenCode := token & mask
			if sentinel >= 0 && lenCode == sentinel {
				if src >= len(payload) {
					return nil
				}
				lenCode = int(payload[src]) + sentinel
				src++
			}

			length := lenCode + 3
			if remaining < length || dst+length > len(out) {
				return nil
			}

			if distance == 0 {
				dst += length
			} else {
				copyFrom := dst - distance
				if copyFrom < 0 {
					return nil
				}
				for i := 0; i < length; i++ {
					out[dst+i] = out[copyFrom+i]
				}
				dst += length
			}
			remaining -= length
		}

		flags >>= 1
	}
	return out
}
