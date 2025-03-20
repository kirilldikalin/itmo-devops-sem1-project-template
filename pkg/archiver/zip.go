package archiver

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"path/filepath"
)

type ZipArchiver struct{}

func NewZipArchiver() *ZipArchiver {
	return &ZipArchiver{}
}

func (z *ZipArchiver) Extract(input io.Reader) (io.ReadCloser, error) {
	data, err := io.ReadAll(input)
	if err != nil {
		return nil, err
	}
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	for _, file := range zipReader.File {
		if filepath.Ext(file.Name) == ".csv" {
			f, err := file.Open()
			if err != nil {
				return nil, err
			}
			content, err := io.ReadAll(f)
			f.Close()
			if err != nil {
				return nil, err
			}
			return io.NopCloser(bytes.NewReader(content)), nil
		}
	}
	return nil, errors.New("no csv file found in the provided archive")
}

func (z *ZipArchiver) Archive(output io.Writer, fileName string, data []byte) error {
	zipWriter := zip.NewWriter(output)
	defer zipWriter.Close()

	f, err := zipWriter.Create(fileName)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	return err
}
