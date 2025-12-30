package buffer

import "bytes"

func Flush(buf *bytes.Buffer) (string, bool) {
	if buf.Len() == 0 {
		return "", false
	}
	content := buf.String()

	if len(content) > 0 {
		buf.Reset()
		return content, true
	}
	return "", false

}
