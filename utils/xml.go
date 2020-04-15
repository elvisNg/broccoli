package utils

import (
    "encoding/xml"
    "io"
)

func XMLMarshal(v interface{}) ([]byte, error) {
    return xml.Marshal(v)
}

//MarshalIndent MarshalIndent
func XMLMarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
    return xml.MarshalIndent(v, prefix, indent)
}

// Unmarshal、
func XMLUnmarshal(data []byte, v interface{}) error {
    return xml.Unmarshal(data, v)
}

// NewEncoder returns a new encoder that writes to w.
func XMLNewEncoder(w io.Writer) *xml.Encoder {
    return xml.NewEncoder(w)
}

// NewDecoder returns a new decoder that reads from r.
func XMLNewDecoder(r io.Reader) *xml.Decoder {
    return xml.NewDecoder(r)
}
