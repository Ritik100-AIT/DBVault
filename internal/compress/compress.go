package compress

import "io"

func GzipCompress(src io.Reader) (io.Reader, error) {
	return src, nil
}

func GzipDecompress(src io.Reader) (io.Reader, error) {
	return src, nil
}
