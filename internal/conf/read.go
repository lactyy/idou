package conf

import (
	"encoding/json"
	"fmt"
	"github.com/pelletier/go-toml"
	"io"
	"os"
	"path/filepath"
)

func Read[T any](encoding Encoding, path string, d T) (t T, err error) {
	if filepath.Ext(path) == "" {
		path += encoding.Ext()
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return t, err // The operation above is already wrapped by os.PathError
		}
		defer f.Close()
		if err := encoding.Encode(f, d); err != nil {
			return t, fmt.Errorf("encode default: %w", err)
		}
		return d, nil
	} else if err != nil {
		return t, fmt.Errorf("stat: %w", err)
	}
	f, err := os.Open(path)
	if err != nil {
		return t, nil // The operation above is already wrapped by os.PathError
	}
	defer f.Close()
	if err := encoding.Decode(f, &t); err != nil {
		return t, fmt.Errorf("decode: %w", err)
	}
	return t, nil
}

type Encoding interface {
	Encode(w io.Writer, v any) error
	Decode(r io.Reader, v any) error

	Ext() string
}

var (
	TOMLEncoding Encoding = tomlEncoding{}
	JSONEncoding Encoding = jsonEncoding{}
)

type tomlEncoding struct{}

func (tomlEncoding) Encode(w io.Writer, v any) error {
	return toml.NewEncoder(w).Encode(v)
}

func (tomlEncoding) Decode(r io.Reader, v any) error {
	return toml.NewDecoder(r).Decode(v)
}

func (tomlEncoding) Ext() string { return ".toml" }

type jsonEncoding struct{}

func (jsonEncoding) Encode(w io.Writer, v any) error {
	return json.NewEncoder(w).Encode(v)
}

func (jsonEncoding) Decode(r io.Reader, v any) error {
	return json.NewDecoder(r).Decode(v)
}

func (jsonEncoding) Ext() string { return ".json" }
