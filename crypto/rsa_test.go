package crypto

import (
	"testing"
	"time"
)

var (
	sss = `{"token": "604213a4-9e4e-11ee-8e42-0242ac110004", "ts": 1711704865}`
	ddd = `lcQu7bxnV3TYOq+OfZB6FA5yDsKgIh99VJKv0yZa7lhT+RGZMToE3zd4Rp7xOmWpRz+AJdbpjeGu5Zm03ylKtvpPZc3h0fidrSilOshekFbowosvsNViBhyUs+iTCpFIj+CKGKnoX2NOWRaznRIOvNHMr6XEMsKcA/9b+YDP9WnzNS0GQslEctGpQOjqqtcKRKs1hDor3BThyMWAGAsuplG81PwIgpntZqrq8C907HhoEThRokojlLmG5tpdam46/S1PhRARfRnlWOvWQeRcB/ncHhscPsffVdldsilckiXRhEr1I/A90JoK73S7f5nmhcTBILxvky41hwtVLodbOQ==`
)

// func RSAGenKey(bits int) error {
// 	/*
// 		生成私钥
// 	*/
// 	//1、使用RSA中的GenerateKey方法生成私钥
// 	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
// 	if err != nil {
// 		return err
// 	}
// 	// 2、通过X509标准将得到的RAS私钥序列化为：ASN.1 的DER编码字符串
// 	privateStream := x509.MarshalPKCS1PrivateKey(privateKey)
// 	// 3、将私钥字符串设置到pem格式块中
// 	block1 := pem.Block{
// 		Type:  "private key",
// 		Bytes: privateStream,
// 	}
// 	// 4、通过pem将设置的数据进行编码，并写入磁盘文件
// 	fPrivate, err := os.Create("privateKey.pem")
// 	if err != nil {
// 		return err
// 	}
// 	defer fPrivate.Close()
// 	err = pem.Encode(fPrivate, &block1)
// 	if err != nil {
// 		return err
// 	}

// 	/*
// 		生成公钥
// 	*/
// 	publicKey := privateKey.PublicKey
// 	publicStream, _ := x509.MarshalPKIXPublicKey(&publicKey)
// 	// publicStream:=x509.MarshalPKCS1PublicKey(&publicKey)
// 	block2 := pem.Block{
// 		Type:  "public key",
// 		Bytes: publicStream,
// 	}
// 	fPublic, err := os.Create("publicKey.pem")
// 	if err != nil {
// 		return err
// 	}
// 	defer fPublic.Close()
// 	pem.Encode(fPublic, &block2)
// 	return nil
// }

func TestRSaAa(t *testing.T) {
	println(time.Now().Format("2006-01-02 15:04:05.000000000"))
	// RSAGenKey(4096)
	// d := NewRSA()
	c := NewRSA()
	// c.SetPublicKey("MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAucwKlVo41ZZxnSRT396V3PoAOWuGX1MrqXvIE3UxKFt1Xd3rIB8YLXZnxsUgxAKau6KVHUhi2ymeHC2ZyQyQSADFgLQwWHgEONmF5QG3xRcPZAMNUq6pOYR0TfHVCPxpRCa64blagevE495XjAAr5ZLR35Kjdpi9Je4KpJlfQDKTr/k/pNKLyxJUG0aKVGv28aiWzc4SAk0YJGS0QXnD0aY6nTrPJ/CninB/wnIUaKFO9sRGAoxHx7cs/46qwQkFqt2hQalfezvEmVW4on//tGKP8aqCk4ak4mcSswpRB6sa1Uk4/tMqQxfCaM9VlUTdOxBkJZ/gPa3j8H4144zdgQIDAQAB")
	c.SetPrivateKey("MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQC5zAqVWjjVlnGdJFPf3pXc+gA5a4ZfUyupe8gTdTEoW3Vd3esgHxgtdmfGxSDEApq7opUdSGLbKZ4cLZnJDJBIAMWAtDBYeAQ42YXlAbfFFw9kAw1Srqk5hHRN8dUI/GlEJrrhuVqB68Tj3leMACvlktHfkqN2mL0l7gqkmV9AMpOv+T+k0ovLElQbRopUa/bxqJbNzhICTRgkZLRBecPRpjqdOs8n8KeKcH/CchRooU72xEYCjEfHtyz/jqrBCQWq3aFBqV97O8SZVbiif/+0Yo/xqoKThqTiZxKzClEHqxrVSTj+0ypDF8Joz1WVRN07EGQln+A9rePwfjXjjN2BAgMBAAECggEABCP4bmii0Ju4L3TSS7BlrZWCsMTlKzWqyO2hwVFAxbH4FR3vcflPbB/x4xuchdG7CghvA0aMMW8lf2JCxZi6lGgz/pDFbQtLqMqsMbTOmB1R8fwhbWDIE6iQgPYtNbSOUf789i/PxQpwilV3pP6R+91AQRe+/dMcj/5UjWN/nGoaYcftEpUxJyqfcU2vq7a4lJABaI1EfDAY0AA3F6NKHo1U62xyjlowtgDVhDQOqwZQe6GdCDduyUnvT0jg+mswArX5/CqbpuhKmYeanvOf5OWn7pT/HR7LwI85FaA0Jjd14SogwszQ7vzL7ldcK1FMj8iRtEX03Hwy2DcbLTiZ4QKBgQDrWlv50omYBlwD77Hw29uwMOKBjq8s6pfG2pVphXd4l4plLmv+dSaH2yH4jhvFqwOucqG6kZEV93CN0rCrXvV7hGr17qSoDHUGFdySWuI0dNg9nsWBjYEpPk1nmSPaSLZlCRLTlWRqCewQ4dBs6cgX0q1iy1wzXtNJgP4n3F26CQKBgQDKGLzC+LaM0JVA2I/Io7O4Mt4aDolfXCRFcIrT6y4+omkM897bbPvklKfrMk7337YT3IxMi3R9UkvzXCkPXoqoDbxA5x2jcqR5qK9ff+05Ul/YdfasV3S41wYv0QUXDkCAwLrsN3avbj2smGKkiqCjNpVmsYbfXLI7AmVBaoVFuQKBgQCFEecVPrw7b8URGLLUi4sQeONo+4MCc3Xzol1+d09QqOZARVocWqK5h+YSQk9jmUkQlpHpCistb2V9WtY3Xw6PkxRjD1acCccU9MFtIuPpYvNtC9uCX77a0fY0EDtcTWaLg/DYHwzSg9+sv/D308sl5SHRzUfJZ+ExGzUY1plL+QKBgQC4TCjPkKpA7fI+SX+N1COPpeu/TWRfLxLwtDoWGdF5UviD1F12MwPfJuWe6aj0CPHtWOIk58PaiVMz4eab2naN3MDBW0I/DMwLGhab+3hlHsbDCohiD/skmQpOTsnahaezAo8z3TyBrQHXRLwoGzd0v9Es7lX1mX37rCqUpkRVGQKBgDDFptCZPtXCV8b5kM38ZPvxXMKoP1ts+u2YWvW+W762TPC88PiWl/IhFd5izBBVVOL8tJArWi7m0B1sMById5yPz6XGNQuGyCQlT3XXKyUK7GKZP9e2Q0WbqKQYPrzaFjWzaRUR9PrgEOStiz5Oj0XV4Or4YStyVYipbMDUlw/N")
	// c.GenerateKey(RSA2048)
	// sss := "1267312shfskdfadfaf" // toolbox.GetRandomString(30002, true) // "1267312shfskdfadfaf"
	s, err := c.DecodeBase64(ddd)
	if err != nil {
		println(err.Error())
		return
	}
	println(s)
	// bb := []byte(sss)
	// x, err := c.Encode(bb)
	// if err != nil {
	// 	println(err.Error())
	// }
	// println("base64: ", x.Base64String(), "\n")
	// println("hex: ", x.HexString(), "\n")
	// println("raw bytes: ", string(x.Bytes()))
	// w := sync.WaitGroup{}
	// w.Add(20000)
	// for i := 0; i < 20000; i++ {
	// 	go func() {
	// 		defer w.Done()
	// 		x, err := d.Encode(bb)
	// 		if err != nil {
	// 			println(err.Error())
	// 		}
	// 		xs, err := c.Decode(x.Bytes())
	// 		if err != nil {
	// 			println(err.Error())
	// 		}
	// 		if xs != sss {
	// 			println("encode decode not match")
	// 		}
	// 	}()
	// }
	// w.Wait()
	println("done")
}

func TestSign(t *testing.T) {
	sss := "1267312shfskdfadfaf" // toolbox.GetRandomString(30002, true) // "1267312shfskdfadfaf"
	c := NewRSA()
	c.SetPublicKeyFromFile("publicKey.pem")
	c.SetPrivateKeyFromFile("privateKey.pem")
	x, err := c.Sign([]byte(sss))
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	println(x.HexString())
	z, err := c.VerifySign(x.Bytes(), []byte(sss))
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if !z {
		t.Fail()
		return
	}
	z, err = c.VerifySignFromBase64(x.Base64String(), []byte(sss))
	if err != nil {
		t.Fatal(err.Error())
		return
	}
	if !z {
		t.Fail()
	}
}

func BenchmarkRSA(b *testing.B) {
	c := NewRSA()
	c.GenerateKey(RSA2048)
	// c.SetPublicKeyFromFile("publicKey.pem")
	// c.SetPrivateKeyFromFile("privateKey.pem")
	sss := `{"deviceCode":"2","kIndexes":[{"loopIndex":"K1","ctlTimes":[{"date":"2025-01-16","times":[{"period":1,"startTime":"2025-01-16 17:44","endTime":"2025-01-17 07:03"}]},{"date":"2025-01-17","times":[{"period":1,"startTime":"2025-01-17 17:44","endTime":"2025-01-18 07:03"}]}]},{"loopIndex":"K2","ctlTimes":[{"date":"2025-01-16","times":[{"period":1,"startTime":"2025-01-16 17:44","endTime":"2025-01-17 07:03"}]},{"date":"2025-01-17","times":[{"period":1,"startTime":"2025-01-17 17:44","endTime":"2025-01-18 07:03"}]}]},{"loopIndex":"K3","ctlTimes":[{"date":"2025-01-16","times":[{"period":1,"startTime":"2025-01-16 17:44","endTime":"2025-01-17 07:03"}]},{"date":"2025-01-17","times":[{"period":1,"startTime":"2025-01-17 17:44","endTime":"2025-01-18 07:03"}]},{"loopIndex":"K4","ctlTimes":[{"date":"2025-01-16","times":[{"period":1,"startTime":"2025-01-16 17:44","endTime":"2025-01-17 07:03"}]},{"date":"2025-01-17","times":[{"period":1,"startTime":"2025-01-17 17:44","endTime":"2025-01-18 07:03"}]}]},{"loopIndex":"K5","ctlTimes":[{"date":"2025-01-16","times":[{"period":1,"startTime":"2025-01-16 17:44","endTime":"2025-01-17 07:03"}]},{"date":"2025-01-17","times":[{"period":1,"startTime":"2025-01-17 17:44","endTime":"2025-01-18 07:03"}]}]},{"loopIndex":"K6","ctlTimes":[{"date":"2025-01-16","times":[{"period":1,"startTime":"2025-01-16 17:44","endTime":"2025-01-17 07:03"}]},{"date":"2025-01-17","times":[{"period":1,"startTime":"2025-01-17 17:44","endTime":"2025-01-18 07:03"}]}]},{"loopIndex":"K7","ctlTimes":[{"date":"2025-01-16","times":[{"period":1,"startTime":"2025-01-16 17:44","endTime":"2025-01-17 07:03"}]},{"date":"2025-01-17","times":[{"period":1,"startTime":"2025-01-17 17:44","endTime":"2025-01-18 07:03"}]}]},{"loopIndex":"K8","ctlTimes":[{"date":"2025-01-16","times":[{"period":1,"startTime":"2025-01-16 17:44","endTime":"2025-01-17 07:03"}]},{"date":"2025-01-17","times":[{"period":1,"startTime":"2025-01-17 17:44","endTime":"2025-01-18 07:03"}]}]}]}]}` // toolbox.GetRandomString(30002, true) // "1267312shfskdfadfaf"
	bb := []byte(sss)
	var err error
	var x CValue
	// var xs string
	x, _ = c.Encode(bb)
	bbb := x.Bytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = c.Decode(bbb)
		if err != nil {
			b.Fatal(err.Error())
		}
	}
}
