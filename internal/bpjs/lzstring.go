package bpjs

import (
	"encoding/binary"
	"fmt"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// Implementasi algoritma LZString (pieroxy/lz-string) varian
// compressToEncodedURIComponent / decompressFromEncodedURIComponent,
// identik dengan yang dipakai BPJS Kesehatan untuk kompresi payload.
//
// Semantik mengikuti JavaScript: seluruh operasi berlangsung pada
// unit kode UTF-16 (charCodeAt/charAt), sehingga konversi ke rune
// hanya dilakukan satu kali di batas fungsi.

const uriSafeAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+-$"

var uriSafeIndex = func() map[rune]uint32 {
	m := make(map[rune]uint32)
	for i, r := range uriSafeAlphabet {
		m[r] = uint32(i)
	}
	return m
}()

func utf16Units(s string) ([]uint16, error) {
	if !utf8.ValidString(s) {
		s = strings.ToValidUTF8(s, string(utf8.RuneError))
	}
	return utf16.Encode([]rune(s)), nil
}

func unitsToString(units []uint16) string {
	return string(utf16.Decode(units))
}

// u16key mengubah slice unit UTF-16 menjadi kunci map tanpa merusak
// surrogate pair (repraesentasi byte mentah, hanya untuk perbandingan).
func u16key(u []uint16) string {
	b := make([]byte, len(u)*2)
	for i, v := range u {
		binary.LittleEndian.PutUint16(b[i*2:], v)
	}
	return string(b)
}

// CompressToEncodedURIComponent mengompresi string ke format URI-safe.
// Mengembalikan string kosong untuk input kosong (perilaku JS).
func CompressToEncodedURIComponent(input string) string {
	if input == "" {
		return ""
	}

	units, _ := utf16Units(input)
	const bitsPerChar = 6

	var out []uint16
	var dataVal uint32
	dataPos := 0

	emit := func() {
		out = append(out, uint16(uriSafeAlphabet[dataVal]))
		dataVal = 0
	}

	write := func(value uint32, count int) {
		for i := 0; i < count; i++ {
			dataVal = (dataVal << 1) | (value & 1)
			if dataPos == bitsPerChar-1 {
				dataPos = 0
				emit()
			} else {
				dataPos++
			}
			value >>= 1
		}
	}

	dictionary := make(map[string]uint32)
	toCreate := make(map[string]bool)

	w := []uint16{}
	enlargeIn := uint32(2)
	dictSize := uint32(3)
	numBits := 2

	for _, cu := range units {
		cKey := u16key([]uint16{cu})
		if _, ok := dictionary[cKey]; !ok {
			dictionary[cKey] = dictSize
			dictSize++
			toCreate[cKey] = true
		}

		wc := append(append([]uint16{}, w...), cu)
		wcKey := u16key(wc)
		if _, ok := dictionary[wcKey]; ok {
			w = wc
			continue
		}

		wKey := u16key(w)
		if toCreate[wKey] {
			firstCode := uint32(w[0])
			if firstCode < 256 {
				write(0, numBits)
				write(firstCode, 8)
			} else {
				write(1, numBits)
				write(firstCode, 16)
			}
			enlargeIn--
			if enlargeIn == 0 {
				enlargeIn = 1 << numBits
				numBits++
			}
			delete(toCreate, wKey)
		} else {
			write(dictionary[wKey], numBits)
		}

		enlargeIn--
		if enlargeIn == 0 {
			enlargeIn = 1 << numBits
			numBits++
		}

		dictionary[wcKey] = dictSize
		dictSize++
		w = []uint16{cu}
	}

	if len(w) > 0 {
		wKey := u16key(w)
		if toCreate[wKey] {
			firstCode := uint32(w[0])
			if firstCode < 256 {
				write(0, numBits)
				write(firstCode, 8)
			} else {
				write(1, numBits)
				write(firstCode, 16)
			}
			enlargeIn--
			if enlargeIn == 0 {
				enlargeIn = 1 << numBits
				numBits++
			}
			delete(toCreate, wKey)
		} else {
			write(dictionary[wKey], numBits)
		}
		enlargeIn--
		if enlargeIn == 0 {
			enlargeIn = 1 << numBits
			numBits++
		}
	}

	// Penanda akhir stream (kode 2).
	write(2, numBits)

	// Flush sisa bit.
	for {
		dataVal <<= 1
		if dataPos == bitsPerChar-1 {
			emit()
			break
		}
		dataPos++
	}

	return unitsToString(out)
}

// DecompressFromEncodedURIComponent mendekompresi output
// CompressToEncodedURIComponent kembali ke string aslinya.
func DecompressFromEncodedURIComponent(compressed string) (string, error) {
	if compressed == "" {
		return "", nil
	}

	units, err := utf16Units(compressed)
	if err != nil {
		return "", err
	}

	getNext := func(idx int) (uint32, error) {
		if idx >= len(units) {
			return 0, fmt.Errorf("indeks %d di luar jangkauan", idx)
		}
		v, ok := uriSafeIndex[rune(units[idx])]
		if !ok {
			return 0, fmt.Errorf("karakter tidak valid pada posisi %d", idx)
		}
		return v, nil
	}

	resetValue := 32 // mask awal: 100000 → mengekstrak 6 bit per karakter
	pos := resetValue
	idx := 1
	val, err := getNext(0)
	if err != nil {
		return "", err
	}

	read := func(n int) (uint32, error) {
		var bits uint32
		maxpower := uint32(1) << n
		power := uint32(1)
		for power != maxpower {
			resb := val & uint32(pos)
			pos >>= 1
			if pos == 0 {
				pos = resetValue
				val, err = getNext(idx)
				if err != nil {
					return 0, err
				}
				idx++
			}
			if resb > 0 {
				bits |= power
			}
			power <<= 1
		}
		return bits, nil
	}

	unitFromCode := func(code uint32) []uint16 { return []uint16{uint16(code)} }

	dictionary := make(map[uint32][]uint16)
	for i := 0; i < 3; i++ {
		dictionary[uint32(i)] = []uint16{uint16(i)}
	}

	enlargeIn := 4
	dictSize := 4
	numBits := 3

	next, err := read(2)
	if err != nil {
		return "", err
	}

	var first []uint16
	switch next {
	case 0:
		code, err := read(8)
		if err != nil {
			return "", err
		}
		first = unitFromCode(code)
	case 1:
		code, err := read(16)
		if err != nil {
			return "", err
		}
		first = unitFromCode(code)
	default:
		return "", nil
	}

	dictionary[3] = first
	w := first
	result := [][]uint16{first}

	for {
		c, err := read(numBits)
		if err != nil {
			return "", err
		}

		var entry []uint16
		switch c {
		case 0:
			code, err := read(8)
			if err != nil {
				return "", err
			}
			entry = unitFromCode(code)
			dictionary[uint32(dictSize)] = entry
			dictSize++
			enlargeIn--
		case 1:
			code, err := read(16)
			if err != nil {
				return "", err
			}
			entry = unitFromCode(code)
			dictionary[uint32(dictSize)] = entry
			dictSize++
			enlargeIn--
		case 2:
			total := 0
			for _, part := range result {
				total += len(part)
			}
			joined := make([]uint16, 0, total)
			for _, part := range result {
				joined = append(joined, part...)
			}
			return unitsToString(joined), nil
		default:
			e, ok := dictionary[c]
			switch {
			case ok:
				entry = e
			case int(c) == dictSize:
				entry = append(append([]uint16{}, w...), w[0])
			default:
				return "", fmt.Errorf("kode kamus tidak valid: %d", c)
			}
		}

		if enlargeIn == 0 {
			enlargeIn = 1 << numBits
			numBits++
		}

		result = append(result, entry)
		dictionary[uint32(dictSize)] = append(append([]uint16{}, w...), entry[0])
		dictSize++
		enlargeIn--

		w = entry
		if enlargeIn == 0 {
			enlargeIn = 1 << numBits
			numBits++
		}
	}
}
