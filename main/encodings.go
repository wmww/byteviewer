package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"unicode"
	"unicode/utf8"
)

type encoding struct {
	Name        string
	EncoderFunc func([]byte) (string, int, bool)
	Enabled     bool
	ByteLength  int
	// single character separator
	Separator string
	Desc      string
	MaxWidth  int
	buffer    []byte
	total     int
}

var termColors = []string{
	"31",
	"33",
	"32",
	"36",
	"34",
	"35",
}

func (e *encoding) Encode(chunk []byte) string {
	output := make([]string, 0)
	increment := max(e.ByteLength, 1)
	paddingSize := -(len(e.buffer) * (e.MaxWidth + len(e.Separator)) / increment)
	if e.ByteLength > 0 {
		paddingSize = 0
	}
	start := 0
	maxStart := len(e.buffer)
	outputVisibleLen := 0
	e.buffer = append(e.buffer, chunk...)
	end := increment
	if maxStart == 0 {
		return ""
	}
	for start < maxStart && end <= len(e.buffer) {
		encoded, consumed, important := e.EncoderFunc(e.buffer[start:end])
		encodedLen := utf8.RuneCountInString(encoded)
		if encodedLen > 0 {
			paddingSize += e.MaxWidth
			if paddingSize > encodedLen {
				encoded = fmt.Sprintf("%*s%s", paddingSize - encodedLen, "", encoded)
			}
			outputVisibleLen += utf8.RuneCountInString(encoded)
			if (!disableColors) {
				colorIndex := ((e.total + start) / colorWidth) % len(termColors)
				var brightness = "1" // bold
				if !important {
					brightness = "2" // faint
				}
				encoded = "\x1b[" + brightness + ";" + termColors[colorIndex] + "m" + encoded
			}
			paddingSize = 0
			output = append(output, encoded)
		} else {
			paddingSize += e.MaxWidth + len(e.Separator)
		}
		if consumed == 0 {
			end += increment
		} else {
			if enableOffsets {
				start += 1
			} else {
				start += consumed
			}
			end = start + increment
		}
	}
	e.buffer = e.buffer[start:]
	e.total += start
	wdth := e.EncodingWidth(bufferSize)
	outputVisibleLen += (len(output) - 1) * len(e.Separator) // Account for separators
	padding := wdth - outputVisibleLen
	// join with separator
	// return string with padding
	return fmt.Sprintf("%s%*s", strings.Join(output, e.Separator), padding, "")
}

// map unicode control chars to ascii equivalents
func unicodeControlToASCII(unicodeRune rune) rune {
	if !unicode.IsControl(unicodeRune) {
		return unicodeRune
	}
	switch unicodeRune {
	case 0x0000:
		return '␀' // null
	case 0x0007:
		return '␇' // bell
	case 0x0008:
		return '⌫' // backspace
	case 0x0009:
		return '⇥' // tab
	case 0x000A, 0x000B, 0x000C, 0x000D, 0x0085, 0x2028, 0x2029:
		return '⏎' // newline and related
	case 0x001B:
		return '⎋' // escape
	default:
		// return generic sp char
		return '␀'
	}
}

var mapInvalidChar = map[uint8]rune{
	'\n':   '⏎',
	'\t':   '⇥',
	'\r':   '↵',
	'\v':   '↴',
	'\f':   '↵',
	'\b':   '⌫',
	'\a':   '␇',
	'\x1b': '⎋',
	'\x00': '␀',
}

func parseASCII(chunk []byte) (string, int, bool) {
	var output string
	var important = false
	for _, b := range chunk {
		if b >= 32 && b <= 126 { // Printable ASCII range
			important = true
			rn, ok := mapInvalidChar[b]
			if ok {
				output += fmt.Sprintf("%c", rn)
			} else {
				output += fmt.Sprintf("%c", b)
			}
		} else {
			output += "." // Non-printable characters are represented as a dot
		}
	}
	return output, len(chunk), important

}

func utf8GetRune(chunk []byte) (rune, int) {
	if utf8.Valid(chunk) {
		r, _ := utf8.DecodeRune(chunk)
		return r, len(chunk)
	} else {
		for i := 1; i < len(chunk); i++ {
			if utf8.RuneStart(chunk[i]) || i > utf8.UTFMax {
				// Either a new rune has been started without the last one being finished or we've gotten
				// more bytes than fit in a UTF-8 rune.
				// Non-printable characters are represented as U+FFFD (REPLACEMENT CHARACTER)
				return '�', i
			}
		}
	}
	return 0, 0
}

func (e *encoding) EncodingWidth(bytewidth int) int {
	numEntries := bytewidth / max(e.ByteLength, 1)
	if enableOffsets {
		numEntries = bytewidth
	}
	return ((e.MaxWidth * numEntries) + (numEntries - 1) * len(e.Separator)) // separators
}

var encodings = []encoding{
	{
		Name: "i8",
		EncoderFunc: func(b []byte) (string, int, bool) {
			var val = int8(b[0])
			return fmt.Sprintf("%d", val), len(b), true
		},
		Enabled:    false,
		ByteLength: 1,
		Separator:  `,`,
		MaxWidth:   4,
		Desc:       `Signed 8-bit integer`,
	},
	{
		Name: "u8",
		EncoderFunc: func(b []byte) (string, int, bool) {
			var val = uint8(b[0])
			return fmt.Sprintf("%d", val), len(b), true
		},
		Enabled:    false,
		ByteLength: 1,
		Separator:  `,`,
		MaxWidth:   3,
		Desc:       `Unsigned 8-bit integer`,
	},
	{
		Name: "i16",
		EncoderFunc: func(b []byte) (string, int, bool) {
			var val = int16(binary.LittleEndian.Uint16(b))
			return fmt.Sprintf("%d", val), len(b), val < 8192 && val > -8192
		},
		Enabled:    false,
		ByteLength: 2,
		Separator:  `,`,
		MaxWidth:   6,
		Desc:       `Signed 16-bit integer`,
	},
	{
		Name: "u16",
		EncoderFunc: func(b []byte) (string, int, bool) {
			var val = binary.LittleEndian.Uint16(b)
			return fmt.Sprintf("%d", val), len(b), val < 8192
		},
		Enabled:    false,
		ByteLength: 2,
		Separator:  `,`,
		MaxWidth:   6,
		Desc:       `Unsigned 16-bit integer`,
	},
	{
		Name: "i32",
		EncoderFunc: func(b []byte) (string, int, bool) {
			return fmt.Sprintf("%d", int32(binary.LittleEndian.Uint32(b))), len(b), true
		},
		Enabled:    false,
		ByteLength: 4,
		Separator:  `,`,
		MaxWidth:   11,
		Desc:       `Signed 32-bit integer`,
	},
	{
		Name: "u32",
		EncoderFunc: func(b []byte) (string, int, bool) {
			return fmt.Sprintf("%d", binary.LittleEndian.Uint32(b)), len(b), true
		},
		Enabled:    false,
		ByteLength: 4,
		Separator:  `,`,
		MaxWidth:   11,
		Desc:       `Unsigned 32-bit integer`,
	},
	{
		Name: "f32",
		EncoderFunc: func(b []byte) (string, int, bool) {
			return fmt.Sprintf("%.6g", math.Float32frombits(binary.LittleEndian.Uint32(b))), len(b), true
		},
		Enabled:    false,
		ByteLength: 4,
		Separator:  `,`,
		MaxWidth:   12,
		Desc:       `IEEE 754 single-precision binary floating-point format: sign bit, 8 bits exponent, 23 bits mantissa`,
	},
	{
		Name: "i64",
		EncoderFunc: func(b []byte) (string, int, bool) {
			return fmt.Sprintf("%d", int64(binary.LittleEndian.Uint64(b))), len(b), true
		},
		Enabled:    false,
		ByteLength: 8,
		Separator:  `,`,
		MaxWidth:   20,
		Desc:       `Signed 64-bit integer`,
	},
	{
		Name: "u64",
		EncoderFunc: func(b []byte) (string, int, bool) {
			return fmt.Sprintf("%d", binary.LittleEndian.Uint64(b)), len(b), true
		},
		Enabled:    false,
		ByteLength: 8,
		Separator:  `,`,
		MaxWidth:   20,
		Desc:       `Unsigned 64-bit integer`,
	},
	{
		Name: "f64",
		EncoderFunc: func(b []byte) (string, int, bool) {
			return fmt.Sprintf("%.6g", math.Float64frombits(binary.LittleEndian.Uint64(b))), len(b), true
		},
		Enabled:    false,
		ByteLength: 8,
		Separator:  `,`,
		MaxWidth:   12,
		Desc:       `IEEE 754 double-precision binary floating-point format: sign bit, 11 bits exponent, 52 bits mantissa`,
	},
	{
		Name:        "hex",
		EncoderFunc: func(b []byte) (string, int, bool) {
			return fmt.Sprintf("%x", b), len(b), true
		},
		Enabled:     false,
		ByteLength:  1,
		Separator:   `,`,
		MaxWidth:    2,
		Desc:        `Hexadecimal encoding`,
	},
	{
		Name:        "utf8h",
		EncoderFunc: func(b []byte) (string, int, bool) {
			r, consumed := utf8GetRune(b)
			if consumed > 0 {
				return fmt.Sprintf("%x", r), consumed, true
			} else {
				return "", consumed, false
			}
		},
		Enabled:     false,
		ByteLength:  0,
		Separator:   `,`,
		MaxWidth:    3,
		Desc:        `Unicode code points of UTF-8 encoded text. Hexadecimal.`,
	},
	{
		Name:        "utf8i",
		EncoderFunc: func(b []byte) (string, int, bool) {
			r, consumed := utf8GetRune(b)
			if consumed > 0 {
				return fmt.Sprintf("%d", r), consumed, true
			} else {
				return "", consumed, false
			}
		},
		Enabled:     false,
		ByteLength:  0,
		Separator:   `,`,
		MaxWidth:    3,
		Desc:        `Unicode code points of UTF-8 encoded text. Decimal.`,
	},
	{
		Name:        "ascii",
		EncoderFunc: parseASCII,
		Enabled:     false,
		ByteLength:  1,
		Separator:   ``,
		MaxWidth:    1,
		Desc:        `ASCII encoded text. Non-printable characters are represented as a dot and the following characters are replaced with their unicode equivalents: \\n, \\t, \\r, \\v, \\f, \\b, \\a, \\x1b`,
	},
	{
		Name:        "utf8",
		EncoderFunc: func(b []byte) (string, int, bool) {
			r, consumed := utf8GetRune(b)
			if consumed > 0 {
				// replace control chars with unicode equivalents
				return fmt.Sprintf("%c", unicodeControlToASCII(r)), consumed, true
			} else {
				return "", consumed, false
			}
		},
		Enabled:     false,
		ByteLength:  0,
		Separator:   ``,
		MaxWidth:    1,
		Desc:        `UTF-8 encoded text. Replaces control characters with unicode symbol equivalents (mostly).`,
	},
}
