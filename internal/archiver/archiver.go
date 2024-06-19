package archiver

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

func CreateArchive(archive string, files []string) error {
	buf, err := os.Create(archive)
	if err != nil {
		return err
	}
	defer buf.Close()

	gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	for _, file := range files {
		err := addToArchive(tw, filepath.Dir(file), "", filepath.Base(file))
		if err != nil {
			return err
		}
	}

	return nil
}

func ExtractArchive(archive string, output string) error {
	if err := os.MkdirAll(output, 0755); err != nil {
		return err
	}

	buf, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer buf.Close()

	gw, err := gzip.NewReader(buf)
	defer gw.Close()
	tw := tar.NewReader(gw)

	for {
		header, err := tw.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Typeflag == tar.TypeReg {
			folder := filepath.Join(output, filepath.Dir(header.Name))
			if err := os.MkdirAll(folder, 0755); err != nil {
				return err
			}

			data := make([]byte, header.Size)
			_, err := tw.Read(data)
			if err != nil && err != io.EOF {
				return err
			}

			file := filepath.Join(output, header.Name)
			fo, err := os.Create(file)
			if err != nil {
				return err
			}
			defer fo.Close()

			if _, err := fo.Write(data); err != nil {
				return err
			}
		}
	}

	return nil
}

func addToArchive(tw *tar.Writer, base string, prefix string, filename string) error {
	fullname := filepath.Join(base, prefix, filename)
	info, err := os.Stat(fullname)
	if err != nil {
		return err
	}

	file, err := os.Open(fullname)
	if err != nil {
		return err
	}
	defer file.Close()

	if info.IsDir() {
		files, err := file.Readdir(-1)
		if err != nil {
			return err
		}

		for _, file := range files {
			if err = addToArchive(tw, base, filepath.Join(prefix, filename), file.Name()); err != nil {
				return err
			}
		}
	} else {
		header, err := tar.FileInfoHeader(info, info.Name())
		if err != nil {
			return err
		}

		header.Name = filepath.Join(prefix, filename)

		err = tw.WriteHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(tw, file)
		if err != nil {
			return err
		}
	}

	return nil
}
