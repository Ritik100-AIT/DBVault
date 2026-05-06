package compress

import (
	"compress/gzip"
	"io"
)

// GzipCompress compresses the input stream and returns a reader for the compressed data.
func GzipCompress(src io.Reader) (io.Reader, error) {
	pr, pw := io.Pipe()
	gw := gzip.NewWriter(pw)

	go func() {
		defer pw.Close()
		defer gw.Close()
		if _, err := io.Copy(gw, src); err != nil {
			pw.CloseWithError(err)
		}
	}()

	return pr, nil
}

// GzipDecompress wraps the compressed stream with gzip.Reader for reading decompressed bytes.
func GzipDecompress(src io.Reader) (io.Reader, error) {
	gr, err := gzip.NewReader(src)
	if err != nil {
		return nil, err
	}
	return gr, nil
}
