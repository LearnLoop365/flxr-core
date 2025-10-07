package pdfs

import "io"

// Writer — minimal, stream-style, append-only PDF writer. No page navigation
// T: Concrete Template Type -> depends on each implementation
type Writer[T any] interface {
	PaperSize() PaperSize
	Orientation() string

	TemplateStore() *TemplateStore[T]
	ImportPageAsTemplate(filepath string, pageNum int, storeKey string) error

	AddBlankPage()
	AddTemplatePage(storeKey string) error

	SetFont(family string, style string, size float64)

	Text(x float64, y float64, text string)

	WriteTo(w io.Writer) (int64, error)
	WriteToFile(filepath string) error
	ProduceBytes() ([]byte, error)
}

type CountWriter struct {
	w io.Writer
	n int64
}

func NewCountWriter(w io.Writer) *CountWriter {
	return &CountWriter{w: w}
}

// Write implements io.Writer
func (cw *CountWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	cw.n += int64(n) // cuz Write() can be called multiple times internally
	return n, err
}

// BytesWritten returns the total number of bytes written
func (cw *CountWriter) BytesWritten() int64 {
	return cw.n
}
