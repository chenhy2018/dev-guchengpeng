package timeio

import "io"
import "testing"
import "time"

func TestTimeIO(t *testing.T) {

}

type noEndReader struct{}

func (r noEndReader) Read(p []byte) (n int, err error) { return len(p), nil }

type noEndWriter struct{}

func (r noEndWriter) Write(p []byte) (n int, err error) { return len(p), nil }

type sleepWriter struct{}

func (r sleepWriter) Write(p []byte) (n int, err error) {
	time.Sleep(time.Duration(1e5))
	return len(p), nil
}

func BenchmarkTimeIO(b *testing.B) {
	b.StopTimer()
	reader := NewReader(noEndReader{})
	writer := NewWriter(noEndWriter{})
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		io.CopyN(writer, reader, 1024*1024)
	}
}

func BenchmarkTimeIO2(b *testing.B) {
	b.StopTimer()
	reader := NewReader(noEndReader{})
	writer := NewWriter(sleepWriter{})
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		io.CopyN(writer, reader, 1024*1024)
	}
}
