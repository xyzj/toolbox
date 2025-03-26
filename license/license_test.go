package license

import (
	"os"
	"testing"
	"time"
)

const (
	ec = `MHcCAQEEIH21ipUK13z0BpIARIIRxGeklec3lJ/YBQ1MxEm4zBdLoAoGCCqGSM49AwEHoUQDQgAEtK+LlCGyjdOfF72+d4GBRNVYnCxxl5wDUm/pHEcNaOQK1WkiSs2xOb6Ps9a8q6gIP10943b9yxG5gceG4DdLqw==`
	sm = `MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgPcsemvK9+p+3aoTYUXOfmQ3AmUov6F4IbNIZ1t1adH+gCgYIKoEcz1UBgi2hRANCAAT8XxPlls7unXes8I4w33MTagXQ3hfqNbqNOne6oG40yGv6HmuDe7TuORhPYle5Us4Z5kPA05laic5rfcCFt5BB`
)

func TestLic(t *testing.T) {
	x := SignMachine(time.Now().AddDate(50, 0, 0))
	os.WriteFile(".firstrun", []byte(x), 0o664)
	println(x, VerifyMachine(x))
}

func TestAcc(t *testing.T) {
	x := SignAccount("minamoto.xu@outlook.com", time.Now().AddDate(50, 0, 0))
	os.WriteFile(".firstrun", []byte(x), 0o664)
	println(x, VerifyAccount("minamoto.xu@outlook.com", x))
}

func TestVerAcc(t *testing.T) {
	x, _ := os.ReadFile(".firstrun")
	println(VerifyAccount("minamoto.xu@outlook.com", string(x)))
}

func TestVerMac(t *testing.T) {
	x, _ := os.ReadFile(".firstrun")
	println(VerifyMachine(string(x)))
}

func TestEncode(t *testing.T) {
	enecc := encodeAndXOR(ec, 0x19)
	ensm2 := encodeAndXOR(sm, 0x6)
	println(enecc)
	// for _, v := range enecc {
	// 	fmt.Printf("0x%x,", v)
	// }
	println("")
	println(ensm2)

	// for _, v := range ensm2 {
	// 	fmt.Printf("0x%x,", v)
	// }
	deecc := decodeAndXOR([]byte(enecc), 0x19)
	desm2 := decodeAndXOR([]byte(ensm2), 0x06)
	println(deecc == ec, desm2 == sm)
}
