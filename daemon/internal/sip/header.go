package sip

import (
	"fmt"
	"io"
	"net/textproto"
	"strings"
)

type Header map[string][]string

func (h Header) Contains(key string) bool {
	_, ok := h[key]
	return ok
}

func (h Header) Set(key, value string) {
	textproto.MIMEHeader(h).Set(key, value)
}

func (h Header) Add(key, value string) {
	textproto.MIMEHeader(h).Add(key, value)
}

func (h Header) Get(key string) string {
	return textproto.MIMEHeader(h).Get(key)
}

func (h Header) ParseHeader(line string) error {
	parts := strings.Split(line, ":")
	if len(parts) == 1 {
		return fmt.Errorf("%w: invalid header line: %s", ErrParsingError, line)
	}

	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(strings.Join(parts[1:], ":"))

	h.Add(key, val)

	return nil
}

func (h Header) Write(w io.Writer) error {
	for k, vals := range h {
		for _, val := range vals {
			if _, err := io.WriteString(w, k); err != nil {
				return err
			}
			if _, err := io.WriteString(w, ": "); err != nil {
				return err
			}
			if _, err := io.WriteString(w, val); err != nil {
				return err
			}
			if _, err := io.WriteString(w, "\r\n"); err != nil {
				return err
			}
		}
	}
	return nil
}

func (h Header) DebugString() string {
	var b strings.Builder
	h.Write(&b)
	return b.String()
}
