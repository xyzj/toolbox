package crypto

import (
	"testing"
)

func TestHash(t *testing.T) {
	// v := "kjhfksdfh2983u92fsdkfhakjdhf92837@#$^&*()"
	// buf := strings.Builder{}
	v := "aaaaaaaaaaaaa我的aaaaaaa"
	// v := "dlKvmrWa0d5rC7y1rxDhvBw7unmmgUUTFAwMJBnmTbJRnvGgtGvVvkzW5c+vGeiqhBRcQtgJdXpQ47x2OxhkK0Q/AdDSa4LVV/OVHftaYryetcxcLzJEPMPB3i9Eef349BHg0x3nJSvzNXp6qCo5SZFvjQvxivddzXMGuDI6tbT6LTSSM2vObv8ApuDXGBkQr9fc94XzY7TrlIyuAZqVFoWupdbpTOtqlWECVuu03Gwu55/k9bHT6TQDjburgi8mWGCU4e12d51NRw5hAF+eid87B7Q18bEnPs1jEBFce7mDawAawhjeQzpyS4rvETthDXZAnr4+HY5UzPY6PjkVEg=="
	// for _, r := range s {
	// 	if unicode.IsPrint(r) && !unicode.Is(unicode.So, r) {
	// 		buf.WriteRune(r)
	// 	}
	// }
	// println(buf.String(), len(s), len(buf.String()))
	for i := 0; i < 10000; i++ {
		// 	println(ObfuscationString(v))
		// }
		a := ObfuscationString(v)
		b := DeobfuscationString(a)
		if v != b {
			t.Fatalf("not match, %d, %s, %s", i, v, b)
			return
		}
	}

	a := ObfuscationString(v)
	b := DeobfuscationString(a)
	println(b)
}

func BenchmarkObfusNew(b *testing.B) {
	s := "dlKvmrWa0d5rC7y1rxDhvBw7unmmgUUTFAwMJBnmTbJRnvGgtGvVvkzW5c+vGeiqhBRcQtgJdXpQ47x2OxhkK0Q/AdDSa4LVV/OVHftaYryetcxcLzJEPMPB3i9Eef349BHg0x3nJSvzNXp6qCo5SZFvjQvxivddzXMGuDI6tbT6LTSSM2vObv8ApuDXGBkQr9fc94XzY7TrlIyuAZqVFoWupdbpTOtqlWECVuu03Gwu55/k9bHT6TQDjburgi8mWGCU4e12d51NRw5hAF+eid87B7Q18bEnPs1jEBFce7mDawAawhjeQzpyS4rvETthDXZAnr4+HY5UzPY6PjkVEg=="
	for i := 0; i < b.N; i++ {
		ObfuscationString(s)
	}
}

func BenchmarkObfusDeNew(b *testing.B) {
	s := "CxbgU94c3o5PLUZyF62YWA666k89JA6VV7352EUU7k7bxo5K7nZzZ8Y9IR7xsCW5wo5dTC2bNC5Z1UVEVUVS0EZTlE8l6mRAlA7W5l3GJ+3RMBYAJ86n6DWB/85CH8QU+EQIB87FlUW8FkYMI723TRYyV6Wb0CZMACZR9B3RJn0mMo8CVU299U2dhFW0Y7XeFCUnOk4WYl3S3TQRcBU89R3N+UQSuEWk3E8Mv8ZN3k6WQR5SQC053B7HJC729o4TV925f/Ze7l7tVlWTUE2JVCRdVSYSQmXP2UXrKD7VI75eY+Up7BUTjoYPdB3YQnSR+C8SbTRlN+48am8SJB0hHEZ9AUZDZ+5BN95+VSXRJ+4p5mV62m7dZ+PQ6R7VZoRSAQ4qdB2/UCRcTmY7gE29YE6Ed+R1rUZ+QlZ5Sk3OM62CYkQfNS71Z6342RONWn4B/QRVKRRiJ6RnNnVLFS6SHA48MBVQOn3Us+U6GT74IT7a9D4CJC8kFo8kVA6n0oVlV+37T7ZrJB5fL+QI+l06YkZXFm0426XHlEZCZR30slSh9nY"
	for i := 0; i < b.N; i++ {
		DeobfuscationString(s)
	}
}
