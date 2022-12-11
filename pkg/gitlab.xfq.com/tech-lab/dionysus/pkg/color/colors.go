package color

import "io"

type Color []byte

var (
	Green   = Color([]byte{27, 91, 51, 50, 109})
	White   = Color([]byte{27, 91, 51, 55, 109})
	Yellow  = Color([]byte{27, 91, 51, 51, 109})
	Red     = Color([]byte{27, 91, 51, 49, 109})
	Blue    = Color([]byte{27, 91, 51, 52, 109})
	Magenta = Color([]byte{27, 91, 51, 53, 109})
	Cyan    = Color([]byte{27, 91, 51, 54, 109})
)

type colorWriter struct {
	c Color
	w io.Writer
}

// WARNING:: Only use this func in dev or testing.
func (c *colorWriter) Write(p []byte) (n int, err error) {
	l, err := c.w.Write(append(append(c.c, p...), White...))
	n = l - len(c.c) - len(White)
	return
}

func Writer(w io.Writer, c Color) io.Writer {
	return &colorWriter{
		c: c,
		w: w,
	}
}
