package util

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"bytes"
	"github.com/pkg/errors"
)

// FileWriter :
type FileWriter func(WriterFunc, ...string) error

// WriterFunc :
type WriterFunc func(io.Writer) error

// ByteFunc :
type ByteFunc func(bytes.Buffer) error

// NewTemplateWriter :
func NewTemplateWriter(input interface{}, templatename string, templateString string) WriterFunc {
	return func(writer io.Writer) error {
		tmpl, err := template.New(templatename).Parse(templateString)
		if err != nil {
			return errors.Wrap(err, "Error parsing template")
		}
		err = tmpl.Execute(writer, input)
		if err != nil {
			return errors.Wrap(err, "Error processing template")
		}
		return nil
	}
}

// NewStringWriter :
func NewStringWriter(input interface{}, content string) WriterFunc {
	return func(writer io.Writer) error {
		tmpl := bytes.NewBufferString(content)
		_, err := tmpl.WriteTo(writer)
		if err != nil {
			return errors.Wrap(err, "Error processing template")
		}
		return nil
	}
}

// NewFileWriter :
func NewFileWriter(targetFolder string) FileWriter {
	return func(writerFunc WriterFunc, elem ...string) error {
		elem = append(elem, "")
		copy(elem[1:], elem[0:])
		elem[0] = targetFolder
		fp := filepath.Join(elem...)
		os.MkdirAll(path.Dir(fp), os.ModeDir|0755)
		fileToWriteTo, err := os.Create(fp)
		if err != nil {
			return errors.Wrapf(err, "Error creating %+t", elem)
		}
		defer fileToWriteTo.Close()
		err = writerFunc(fileToWriteTo)
		if err != nil {
			return errors.Wrap(err, "Error error writing data")
		}
		err = fileToWriteTo.Sync()
		if err != nil {
			return errors.Wrapf(err, "Error writing %+t", elem)
		}
		return nil
	}
}
