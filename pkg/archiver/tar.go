package archiver

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"path/filepath"
)

type TarArchiver struct{}

func NewTarArchiver() *TarArchiver {
	return &TarArchiver{}
}

func (t *TarArchiver) Extract(input io.Reader) (io.ReadCloser, error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}
	tarReader := tar.NewReader(bytes.NewReader(data))

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if filepath.Ext(header.Name) == ".csv" {
			var buf bytes.Buffer
			if _, err := io.Copy(&buf, tarReader); err != nil {
				return nil, err
			}
			return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
		}
	}
	return nil, errors.New("no csv file found in the provided archive")
}

func (t *TarArchiver) Archive(output io.Writer, fileName string, data []byte) error {
	tw := tar.NewWriter(output)
	defer tw.Close()

	const fileMode0600 = 0600

	header := &tar.Header{
		Name: fileName,
		Mode: fileMode0600,
		Size: int64(len(data)),
	}
	if err := tw.WriteHeader(header); err != nil {
		return err
	}
	_, err := tw.Write(data)
	return err
}
