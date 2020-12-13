package build

import (
	"bytes"
	"io"
	"time"
)

func send(w io.Writer, msg []byte) error {
	_, err := w.Write(msg)
	return err
}

func respond(w io.Writer, msg string) error {

	buff := &bytes.Buffer{}

	buff.WriteString(time.Now().Format("2006/01/02 15:04:05.000000000 MST"))
	buff.WriteString(": ")

	buff.WriteString(msg)

	buff.WriteString("\n")

	return send(w, buff.Bytes())
}

