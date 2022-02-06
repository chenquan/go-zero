//go:build sonic
// +build sonic

package jsonx

import (
	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/decoder"
	"io"
)

// Marshal marshals v into json bytes.
var Marshal = sonic.Marshal

// Unmarshal unmarshals data bytes into v.
func Unmarshal(data []byte, v interface{}) error {
	s := string(data)
	decode := decoder.NewDecoder(s)
	if err := unmarshalUseNumber(decode, v); err != nil {
		return formatError(s, err)
	}

	return nil
}

// UnmarshalFromString unmarshals v from str.
func UnmarshalFromString(str string, v interface{}) error {
	decode := decoder.NewDecoder(str)
	if err := unmarshalUseNumber(decode, v); err != nil {
		return formatError(str, err)
	}

	return nil
}

// UnmarshalFromReader unmarshals v from reader.
func UnmarshalFromReader(reader io.Reader, v interface{}) error {
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	s := string(data)
	decode := decoder.NewDecoder(s)
	if err := unmarshalUseNumber(decode, v); err != nil {
		return formatError(s, err)
	}

	return nil
}

func unmarshalUseNumber(decoder *decoder.Decoder, v interface{}) error {
	decoder.UseNumber()
	return decoder.Decode(v)
}
