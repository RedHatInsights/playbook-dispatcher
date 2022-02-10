package utils

import (
	"bufio"
	"io"
)

type Compression string

const (
	GZip Compression = "gzip"
	XZ   Compression = "xz"
)

const (
	// https://tools.ietf.org/html/rfc1952#section-2.3.1
	gzipByte1 = 0x1f
	gzipByte2 = 0x8b
	// https://tukaani.org/xz/xz-file-format.txt
	// Section 2.1.1.1.
	xzByte1 = 0xfd
	xzByte2 = 0x37
	xzByte3 = 0x7a
	xzByte4 = 0x58
	xzByte5 = 0x5a
	xzByte6 = 0x00
)

func GetCompressionType(reader io.Reader) (Compression, error) {
	bufferedReader := bufio.NewReaderSize(reader, 6)
	peek, err := bufferedReader.Peek(2)

	if err != nil {
		return "", err
	}

	if peek[0] == gzipByte1 && peek[1] == gzipByte2 {
		return GZip, nil
	}
	if peek[0] == xzByte1 && peek[1] == xzByte2 {
		peek, err = bufferedReader.Peek(6)
		if err != nil {
			return "", err
		}
		if peek[2] == xzByte3 && peek[3] == xzByte4 && peek[4] == xzByte5 && peek[5] == xzByte6 {
			return XZ, nil
		}
	}

	return "", nil
}
