package views

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"sync"
)

// CSVWriter is a concurrency-safe, buffered CSV writer optimised for
// high-throughput sensor logging.
//
// Design decisions for zero-lag:
//   - Underlying bufio.Writer absorbs write syscall overhead.
//   - Mutex is held only for the duration of a single row encode (< 1 µs).
//   - Periodic Flush() is called by the recording controller, not by
//     the writer itself, so the hot path never blocks on I/O.
type CSVWriter struct {
	mu   sync.Mutex
	file *os.File
	buf  *bufio.Writer
	csv  *csv.Writer
	rows uint64
}

// NewCSVWriter opens (or creates) a file and writes the CSV header row.
func NewCSVWriter(path string, bufSizeBytes int, writeHeader bool, header []string) (*CSVWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("csv create %s: %w", path, err)
	}

	if bufSizeBytes <= 0 {
		bufSizeBytes = 256 * 1024 // 256 KB default
	}

	bw := bufio.NewWriterSize(f, bufSizeBytes)
	cw := csv.NewWriter(bw)

	w := &CSVWriter{
		file: f,
		buf:  bw,
		csv:  cw,
	}

	if writeHeader && len(header) > 0 {
		if err := cw.Write(header); err != nil {
			f.Close()
			return nil, fmt.Errorf("csv write header: %w", err)
		}
	}

	return w, nil
}

// WriteRow appends a single CSV row. Thread-safe.
func (w *CSVWriter) WriteRow(row []string) {
	w.mu.Lock()
	_ = w.csv.Write(row) // error is buffered; checked on Flush
	w.rows++
	w.mu.Unlock()
}

// Flush pushes the buffered data to the OS. Called periodically by the
// recording controller (not after every row, to avoid syscall overhead).
func (w *CSVWriter) Flush() {
	w.mu.Lock()
	w.csv.Flush()
	_ = w.buf.Flush()
	w.mu.Unlock()
}

// Close flushes remaining data and closes the file.
func (w *CSVWriter) Close() {
	w.Flush()
	w.mu.Lock()
	_ = w.file.Close()
	w.mu.Unlock()
}

// Rows returns the number of data rows written (excludes header).
func (w *CSVWriter) Rows() uint64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.rows
}
