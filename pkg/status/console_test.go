package status
package status

import (
    "testing"
)

func TestBufferedWriterBuffering(t *testing.T) {
    bw := &BufferedWriter{}
    data := []byte("hello world")
    n, err := bw.Write(data)
    if err != nil {
        t.Fatalf("Write returned error: %v", err)
    }
    if n != len(data) {
        t.Fatalf("Write returned len %d, want %d", n, len(data))
    }
    if bw.buf.Len() == 0 {
        t.Fatalf("Expected buffer to contain data after Write")
    }
    bw.Flush()
    if !bw.ready {
        t.Fatalf("Expected writer to be ready after Flush")
    }
    if bw.buf.Len() != 0 {
        t.Fatalf("Expected buffer to be empty after Flush")
    }
}
