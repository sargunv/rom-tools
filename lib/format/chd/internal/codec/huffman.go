package codec

import (
	"fmt"
)

// Huffman decoder for CHD's 8-bit Huffman encoding.
// Based on MAME's huffman.cpp / libchdr implementation.

const (
	huffmanMaxBits = 16 // Maximum code length
)

// HuffmanDecoder decodes Huffman-encoded data.
type HuffmanDecoder struct {
	numCodes uint32
	maxBits  uint8
	lookup   []huffmanLookup // fast lookup table indexed by maxBits-bit prefix
}

type huffmanLookup struct {
	symbol uint16 // decoded symbol
	bits   uint8  // bits consumed
}

// BitReader reads bits from a byte slice (MSB first).
type BitReader struct {
	data   []byte
	bitPos uint32
}

// NewBitReader creates a new BitReader for the given data.
func NewBitReader(data []byte) *BitReader {
	return &BitReader{data: data, bitPos: 0}
}

// ReadBits reads n bits and returns them as a uint32.
func (br *BitReader) ReadBits(n uint32) (uint32, error) {
	if n == 0 {
		return 0, nil
	}
	if n > 32 {
		return 0, fmt.Errorf("cannot read more than 32 bits at once")
	}

	var result uint32
	for i := range n {
		byteIdx := br.bitPos / 8
		bitIdx := 7 - (br.bitPos % 8) // MSB first

		if int(byteIdx) >= len(br.data) {
			return 0, fmt.Errorf("read past end of data at bit %d", br.bitPos)
		}

		bit := (br.data[byteIdx] >> bitIdx) & 1
		result = (result << 1) | uint32(bit)
		br.bitPos++
		_ = i // silence unused warning
	}
	return result, nil
}

// BitsRemaining returns how many bits are left to read.
func (br *BitReader) BitsRemaining() uint32 {
	totalBits := uint32(len(br.data)) * 8
	if br.bitPos >= totalBits {
		return 0
	}
	return totalBits - br.bitPos
}

// NewHuffmanDecoder creates a Huffman decoder for the given number of codes.
func NewHuffmanDecoder(numCodes uint32, maxBits uint8) *HuffmanDecoder {
	return &HuffmanDecoder{
		numCodes: numCodes,
		maxBits:  maxBits,
	}
}

// ImportTreeRLE imports a Huffman tree encoded with RLE.
// This follows libchdr's huffman_import_tree_rle implementation exactly.
func (hd *HuffmanDecoder) ImportTreeRLE(br *BitReader) error {
	// Determine bits per entry based on maxBits
	var numBits uint32
	if hd.maxBits >= 16 {
		numBits = 5
	} else if hd.maxBits >= 8 {
		numBits = 4
	} else {
		numBits = 3
	}

	// Read bit lengths for all symbols
	bitLengths := make([]uint8, hd.numCodes)
	curNode := uint32(0)

	for curNode < hd.numCodes {
		// Read raw value
		nodeBits, err := br.ReadBits(numBits)
		if err != nil {
			return fmt.Errorf("failed to read huffman bits at node %d: %w", curNode, err)
		}

		if nodeBits != 1 {
			// Non-one value is just raw
			bitLengths[curNode] = uint8(nodeBits)
			curNode++
		} else {
			// One is an escape code - read another value
			nodeBits, err = br.ReadBits(numBits)
			if err != nil {
				return fmt.Errorf("failed to read escape value at node %d: %w", curNode, err)
			}

			if nodeBits == 1 {
				// Double 1 means literal 1
				bitLengths[curNode] = 1
				curNode++
			} else {
				// Otherwise read repeat count: the value is repeated (count+3) times
				repCount, err := br.ReadBits(numBits)
				if err != nil {
					return fmt.Errorf("failed to read repeat count at node %d: %w", curNode, err)
				}
				repCount += 3

				if curNode+repCount > hd.numCodes {
					return fmt.Errorf("RLE overflow at node %d: count %d exceeds numCodes %d", curNode, repCount, hd.numCodes)
				}

				for range repCount {
					bitLengths[curNode] = uint8(nodeBits)
					curNode++
				}
			}
		}
	}

	// Build the lookup table from the bit lengths
	return hd.buildFromBitLengths(bitLengths)
}

// buildFromBitLengths builds the lookup table from bit lengths (canonical Huffman).
// Uses the same algorithm as libchdr/MAME for canonical code assignment.
func (hd *HuffmanDecoder) buildFromBitLengths(bitLengths []uint8) error {
	// Find actual maximum bit length used
	actualMaxBits := uint8(0)
	for _, bl := range bitLengths {
		if bl > actualMaxBits {
			actualMaxBits = bl
		}
	}
	if actualMaxBits == 0 {
		// All codes have zero length - create a trivial decoder
		hd.lookup = make([]huffmanLookup, 1)
		return nil
	}
	if actualMaxBits > hd.maxBits {
		actualMaxBits = hd.maxBits
	}

	// Build histogram of bit lengths
	bitHisto := make([]uint32, actualMaxBits+1)
	for _, bl := range bitLengths {
		if bl > 0 && bl <= actualMaxBits {
			bitHisto[bl]++
		}
	}

	// Compute starting codes for each length using libchdr's algorithm
	// (iterates from longest to shortest)
	curStart := uint32(0)
	startCodes := make([]uint32, actualMaxBits+1)
	for codeLen := int(actualMaxBits); codeLen > 0; codeLen-- {
		nextStart := (curStart + bitHisto[codeLen]) >> 1
		startCodes[codeLen] = curStart
		curStart = nextStart
	}

	// Assign codes to symbols (in symbol order, using next available code for each length)
	nextCode := make([]uint32, actualMaxBits+1)
	copy(nextCode, startCodes)

	symbolCodes := make([]uint32, hd.numCodes)
	for i := range hd.numCodes {
		bl := bitLengths[i]
		if bl > 0 && bl <= actualMaxBits {
			symbolCodes[i] = nextCode[bl]
			nextCode[bl]++
		}
	}

	// Build lookup table
	tableSize := uint32(1) << actualMaxBits
	hd.lookup = make([]huffmanLookup, tableSize)
	hd.maxBits = actualMaxBits

	for symbol := range hd.numCodes {
		bl := bitLengths[symbol]
		if bl > 0 && bl <= actualMaxBits {
			code := symbolCodes[symbol]
			// Fill all entries that match this code prefix
			fillBits := actualMaxBits - bl
			fillCount := uint32(1) << fillBits
			baseIdx := code << fillBits
			for j := range fillCount {
				idx := baseIdx + j
				if idx < tableSize {
					hd.lookup[idx] = huffmanLookup{
						symbol: uint16(symbol),
						bits:   bl,
					}
				}
			}
		}
	}

	return nil
}

// Decode reads one symbol from the bit reader using the Huffman table.
func (hd *HuffmanDecoder) Decode(br *BitReader) (uint32, error) {
	if len(hd.lookup) == 0 {
		return 0, fmt.Errorf("huffman table not initialized")
	}

	// Peek maxBits bits for lookup
	remaining := br.BitsRemaining()
	peekBits := min(remaining, uint32(hd.maxBits))
	if peekBits == 0 {
		return 0, fmt.Errorf("no bits remaining")
	}

	// Save position
	startPos := br.bitPos

	bits, err := br.ReadBits(peekBits)
	if err != nil {
		return 0, err
	}

	// Shift to align with table size
	if peekBits < uint32(hd.maxBits) {
		bits <<= (uint32(hd.maxBits) - peekBits)
	}

	entry := hd.lookup[bits]
	if entry.bits == 0 {
		return 0, fmt.Errorf("invalid huffman code at bit %d (bits=%d, table size=%d)", startPos, bits, len(hd.lookup))
	}

	// Rewind and consume only the bits we actually used
	br.bitPos = startPos + uint32(entry.bits)

	return uint32(entry.symbol), nil
}

// Huffman decompresses CHD Huffman-encoded data.
func Huffman(data []byte, outputSize int) ([]byte, error) {
	// CHD Huffman uses 8-bit symbols (256 codes)
	hd := NewHuffmanDecoder(256, huffmanMaxBits)
	br := NewBitReader(data)

	// Import the Huffman tree from RLE-encoded data at start of stream
	if err := hd.ImportTreeRLE(br); err != nil {
		return nil, fmt.Errorf("huffman tree import: %w", err)
	}

	// Decode the data
	result := make([]byte, outputSize)
	for i := range outputSize {
		sym, err := hd.Decode(br)
		if err != nil {
			return nil, fmt.Errorf("huffman decode at %d: %w", i, err)
		}
		result[i] = byte(sym)
	}

	return result, nil
}
