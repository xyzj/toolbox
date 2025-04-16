package license

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/xyzj/toolbox"
	"github.com/xyzj/toolbox/crypto"
)

const (
	ec        = `MHcCAQEEIH21ipUK13z0BpIARIIRxGeklec3lJ/YBQ1MxEm4zBdLoAoGCCqGSM49AwEHoUQDQgAEtK+LlCGyjdOfF72+d4GBRNVYnCxxl5wDUm/pHEcNaOQK1WkiSs2xOb6Ps9a8q6gIP10943b9yxG5gceG4DdLqw==`
	sm        = `MIGTAgEAMBMGByqGSM49AgEGCCqBHM9VAYItBHkwdwIBAQQgPcsemvK9+p+3aoTYUXOfmQ3AmUov6F4IbNIZ1t1adH+gCgYIKoEcz1UBgi2hRANCAAT8XxPlls7unXes8I4w33MTagXQ3hfqNbqNOne6oG40yGv6HmuDe7TuORhPYle5Us4Z5kPA05laic5rfcCFt5BB`
	ecfactory = `MHcCAQEEIKifff7D9qCZPQh8yzT2WwXhDVpmbSP5Li7ZJ+yBpZcvoAoGCCqGSM49AwEHoUQDQgAE8NNd19Z8Gn9G+afF9TiUDtswqrIBdtnanwB2BAeJ95Zcv7Xi87ipvNCwYqG4gq1ikDmlO1VgORkxq94DWUMleQ==`
	rsfactory = `MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCvCCqY7V1X3kak2sgcy8/tGlTJB8BJ8DBFLBmA69qFhWaAB3XcLIZGRdEYK8xvL3HisdhMbL39usRf2I+HfOgMU9z0jeHFIBDCVFFAJd72BnAQCQBhU1Ff84nzshunGy2Z2Jzl/+8qLo7E8cr+XrsPP0GcjM/q4J8pj58VPomQgeHwMswvdOnxEEVodatyHxzp7o63yLhbelgGj7pNuYVMqigqeKgj1xpFnWBJR1ifO3igZX92dQnVXyg2sInZC9pMU/LBwPQBKoHllFRZ2csGvdFPkZ+q9F4gEdxGCrLzVncxs2onhbPasNyU8zytXgFN09BDEewAw++FRcjOqxL9AgMBAAECggEANt0R8R6SAC1pqlnzmOcf1TvIML1Pvj1d/AivM9R6Chk/AEmWXX3NVvCicnekHIPcKElDuezp/sLQtBRfJQJ3gAm3fCaPCqma5zbjwv477ZUmAji4GEmz+6YMa0k8VrxzpkIaKBP5pRiz4rLBxMrvyn8y8z3GHGLtMOtWW6PfHd67ZRvk6dfw63qWxL6kIklyxbLE1fuLIK5Uvrdhx/+Xn1Ij9dYaS95pD9USlKlr9IcziVg1fNINuI0ygKavz2UvgQjMLt11OaqncU7bLEos2jqVjSfh6dkuPEPj0tlG49hLn6bKVal/G0kGwJmur+RmzJa8HUq10MDRyD3PODLENwKBgQDE6nukOv8j6loMXaaCqgAqNVKSFP/KHDHzQ37uNZdpdvR/zSOnwnhUM3uq5ZuC1+/Ucr2JwKUQnSEc1DJ9i72/MQTzE/QmJ3BwDv4TuHwHpeg0qxfajDYAZMQu2by/eKw1RR/Z7prSEuC8JjvnOPilDPq3YRsjrRl5hdbV/D5l1wKBgQDjjLmOAl6SgmpZ7NoNdbJHafYsd8+wnR2eaqvZXiVqdvU9ykR+c8SJ/RgdKdkAhHdZ0syOdMkLTPrnqhWC5TEO7B+yVoCN7ncN2DpqCiX85c3ui7iFeVQYfMYgoyEyeqgSIhU8jzswlIyNGwKJyBOzb/dgk3feVcn4d7V9b7ULSwKBgEzZmmFw6OwyZOxYjEiIYkIWx1/dtCpDsLbYy6vZ3Hq7gxZxkA3D1sQ9x3Dd5UA+WAoYsaIWPhVqzWZ75iybIfWBAwZ+7hUJ5VbGcyBtKnKtrB8J+ug/OkATE4GZVpF/Xe34SUL7XQ7JcU3cuho2YMvBcgOpcTcOMlf8BOlEZXHHAoGAEEYNMzHKL7IyBvJgedvz/xV97Jo4+UmTR7QbTDVmeaG+Ukf39A81fCTkp5lJkrbmjj78MCf8BNnhi9XnKfBYPNf4QFndYckvLIdNTi4hn1+UBb5qWOlfcUzjIoxoIGMTSBC18hnQQt4s2x6WZOIxPoEAcSC6zuTVx/ZPvYPSr9cCgYEAueY8G6xOy2sJw+pOr/PDT1vYc729yr4QBf3lm/mtVTU6xEmNCsr0ocxWJSZPDJuOBbnnY3NoLXOocig/xT1VNcqSYkyh27z1wGGeZp+by2dyY+jFzdX3hsJQ3/CG0wEo8qlxKRYjlk5Oy21SV92ytn1pnWIIOFoaeJBQuOEk1s8=`
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

func TestGen(t *testing.T) {
	e := crypto.NewRSA()
	e.GenerateKey(crypto.RSA2048)
	e.ToFile("", "rsa_gofactory.pem")
}

func TestEncode(t *testing.T) {
	enecc := encodeAndXOR(ec, 0x19)
	ensm2 := encodeAndXOR(sm, 0x6)
	enecf := encodeAndXOR(rsfactory, 0x31)
	println(enecc)
	// for _, v := range enecc {
	// 	fmt.Printf("0x%x,", v)
	// }
	println("")
	println(ensm2)
	println("\nenecf", enecf)
	for _, v := range ensm2 {
		fmt.Printf("0x%x,", v)
	}
	println("\n", toolbox.GetRandomString(16, true))
	deecc := decodeAndXOR([]byte(enecc), 0x19)
	desm2 := decodeAndXOR([]byte(ensm2), 0x06)
	println(deecc == ec, desm2 == sm)
}
