package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/xyzj/toolbox/json"
	"github.com/xyzj/toolbox/pathtool"
)

type ECShortName byte

var (
	// ECPrime256v1 as elliptic.P256() and openssl ecparam -name prime256v1
	ECPrime256v1 ECShortName = 1
	// ECSecp384r1 as elliptic.P384() and openssl ecparam -name secp384r1
	ECSecp384r1 ECShortName = 2
)

// ECC ecc算法
type ECC struct {
	signHash *HASH
	pubKey   *ecdsa.PublicKey
	priKey   *ecdsa.PrivateKey
	// pubEcies *ecies.PublicKey
	// priEcies *ecies.PrivateKey
	pubBytes CValue
	priBytes CValue
}

// Keys 返回公钥和私钥
func (w *ECC) Keys() (CValue, CValue) {
	return w.pubBytes, w.priBytes
}

// GenerateKey 创建ecc密钥对
//
//	返回，pubkey，prikey，error
func (w *ECC) GenerateKey(ec ECShortName) (CValue, CValue, error) {
	var p *ecdsa.PrivateKey
	switch ec {
	case ECPrime256v1:
		p, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case ECSecp384r1:
		p, _ = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	}
	txt, err := x509.MarshalECPrivateKey(p)
	if err != nil {
		return []byte{}, []byte{}, err
	}
	w.priBytes = txt
	txt, err = x509.MarshalPKIXPublicKey(&p.PublicKey)
	if err != nil {
		return []byte{}, []byte{}, err
	}
	w.pubBytes = txt
	w.pubKey = &p.PublicKey
	// w.pubEcies = ecies.ImportECDSAPublic(w.pubKey)
	w.priKey = p
	// w.priEcies = ecies.ImportECDSA(p)
	return w.pubBytes, w.priBytes, nil
}

// ToFile 创建ecc密钥到文件
func (w *ECC) ToFile(pubfile, prifile string) error {
	if prifile != "" {
		block := &pem.Block{
			Type:  "EC PRIVATE KEY",
			Bytes: w.priBytes.Bytes(),
		}
		txt := pem.EncodeToMemory(block)
		err := os.WriteFile(prifile, txt, 0o644)
		if err != nil {
			return err
		}
	}
	if pubfile != "" {
		block := &pem.Block{
			Type:  "PUBLIC KEY",
			Bytes: w.pubBytes.Bytes(),
		}
		txt := pem.EncodeToMemory(block)
		err := os.WriteFile(pubfile, txt, 0o644)
		if err != nil {
			return err
		}
	}
	return nil
}

// SetPublicKeyFromFile 从文件获取公钥
func (w *ECC) SetPublicKeyFromFile(keyPath string) error {
	b, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}
	block, _ := pem.Decode(b)
	return w.SetPublicKey(base64.StdEncoding.EncodeToString(block.Bytes))
}

// SetPublicKey 设置base64编码的公钥
func (w *ECC) SetPublicKey(key string) error {
	bb, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return err
	}
	pubKey, err := x509.ParsePKIXPublicKey(bb)
	if err != nil {
		return err
	}
	w.pubKey = pubKey.(*ecdsa.PublicKey)
	// w.pubEcies = ecies.ImportECDSAPublic(w.pubKey)
	w.pubBytes = bb
	return nil
}

// SetPrivateKeyFromFile 从文件获取私钥
func (w *ECC) SetPrivateKeyFromFile(keyPath string) error {
	b, err := os.ReadFile(keyPath)
	if err != nil {
		return err
	}
	block, _ := pem.Decode(b)
	return w.SetPrivateKey(base64.StdEncoding.EncodeToString(block.Bytes))
}

// SetPrivateKey 设置base64编码的私钥
func (w *ECC) SetPrivateKey(key string) error {
	bb, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return err
	}
	priKey, err := x509.ParseECPrivateKey(bb)
	if err != nil {
		if strings.Contains(err.Error(), "use ParsePKCS8PrivateKey instead") {
			priKeypk8, err := x509.ParsePKCS8PrivateKey(bb)
			if err != nil {
				return err
			}
			priKey = priKeypk8.(*ecdsa.PrivateKey)
		} else {
			return err
		}
	}
	w.priKey = priKey
	// w.priEcies = ecies.ImportECDSA(priKey)
	w.priBytes = bb

	if len(w.pubBytes) == 0 {
		// 没有载入国pubkey，生成新的pubkey
		txt, err := x509.MarshalPKIXPublicKey(&priKey.PublicKey)
		if err != nil {
			return err
		}
		w.pubBytes = txt
		w.pubKey = &priKey.PublicKey
		// w.pubEcies = ecies.ImportECDSAPublic(w.pubKey)
	}
	return nil
}

// Deprecated: Encode ecc加密
func (w *ECC) Encode(b []byte) (CValue, error) {
	return EmptyValue, errors.New("not supported")
}

// Deprecated: Decode ecc解密
func (w *ECC) Decode(b []byte) (string, error) {
	return "", errors.New("not supported")
}

// Deprecated: DecodeBase64 从base64字符串解码
func (w *ECC) DecodeBase64(s string) (string, error) {
	b, err := base64.StdEncoding.DecodeString(FillBase64(s))
	if err != nil {
		return "", err
	}
	return w.Decode(b)
}

// Sign 签名
func (w *ECC) Sign(b []byte) (CValue, error) {
	if w.priKey == nil {
		return EmptyValue, fmt.Errorf("no private key found")
	}
	signature, err := ecdsa.SignASN1(rand.Reader, w.priKey, w.signHash.Hash(b).Bytes())
	if err != nil {
		return EmptyValue, err
	}
	return CValue(signature), nil
}

// VerifySign 验证签名
func (w *ECC) VerifySign(signature, data []byte) (bool, error) {
	if w.pubKey == nil {
		return false, fmt.Errorf("no public key found")
	}
	ok := ecdsa.VerifyASN1(w.pubKey, w.signHash.Hash(data).Bytes(), signature)
	return ok, nil
}

// VerifySignFromBase64 验证base64格式的签名
func (w *ECC) VerifySignFromBase64(signature string, data []byte) (bool, error) {
	bb, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, err
	}
	return w.VerifySign(bb, data)
}

// VerifySignFromHex 验证hexstring格式的签名
func (w *ECC) VerifySignFromHex(signature string, data []byte) (bool, error) {
	bb, err := hex.DecodeString(signature)
	if err != nil {
		return false, err
	}
	return w.VerifySign(bb, data)
}

// Deprecated: Decrypt 兼容旧方法，直接解析base64字符串
func (w *ECC) Decrypt(s string) string {
	x, _ := w.DecodeBase64(s)
	return x
}

// Deprecated: Encrypt 加密，兼容旧方法，直接返回base64字符串
func (w *ECC) Encrypt(s string) string {
	x, err := w.Encode(json.Bytes(s))
	if err != nil {
		return ""
	}
	return x.Base64String()
}

// Deprecated: EncryptTo 加密字符串
func (w *ECC) EncryptTo(s string) CValue {
	x, err := w.Encode(json.Bytes(s))
	if err != nil {
		return EmptyValue
	}
	return x
}

// CreateCert 创建基于ecc算法的数字证书，opt.RootKey无效时，会重新创建私钥和根证书
func (w *ECC) CreateCert(opt *CertOpt) error {
	// 处理参数
	if opt == nil {
		opt = &CertOpt{
			DNS: []string{},
			IP:  []string{},
		}
	}
	if opt.OutPut == "" {
		opt.OutPut = pathtool.GetExecDir()
	}
	if len(opt.DNS) == 0 {
		opt.DNS = []string{"localhost"}
	}
	if len(opt.IP) == 0 {
		opt.IP = []string{"127.0.0.1"}
	}
	ips := make([]net.IP, 0, len(opt.IP))
	sort.Slice(opt.IP, func(i, j int) bool {
		return opt.IP[i] < opt.IP[j]
	})
	for _, v := range opt.IP {
		ips = append(ips, net.ParseIP(v))
	}
	// 处理根证书
	var rootDer, txt []byte
	var err error
	var rootCsr *x509.Certificate
	// 检查私钥
	if opt.RootKey != "" {
		w.SetPrivateKeyFromFile(opt.RootKey)
	}
	if w.priKey == nil {
		opt.RootCa = ""
		opt.RootKey = ""
		w.GenerateKey(ECPrime256v1)
	}
	// 创建根证书
	if opt.RootCa != "" {
		b, err := os.ReadFile(opt.RootCa)
		if err == nil {
			p, _ := pem.Decode(b)
			rootCsr, err = x509.ParseCertificate(p.Bytes)
			if err != nil {
				return err
			}
		}
	}
	if rootCsr == nil {
		rootCsr = &x509.Certificate{
			Version:      3,
			SerialNumber: big.NewInt(time.Now().Unix()),
			Subject: pkix.Name{
				Country:    []string{"CN"},
				Province:   []string{"Shanghai"},
				Locality:   []string{"Shanghai"},
				CommonName: "xyzj",
			},
			NotBefore:             time.Now(),
			NotAfter:              time.Now().AddDate(68, 0, 0),
			MaxPathLen:            1,
			BasicConstraintsValid: true,
			IsCA:                  true,
			KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign | x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment,
			// ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		}
	}
	rootDer, err = x509.CreateCertificate(rand.Reader, rootCsr, rootCsr, w.pubKey, w.priKey)
	if err != nil {
		return err
	}
	// 创建服务器证书
	certCsr := &x509.Certificate{
		Version:      3,
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Country:    []string{"CN"},
			Province:   []string{"Shanghai"},
			Locality:   []string{"Shanghai"},
			CommonName: "xyzj",
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(68, 0, 0),
		DNSNames:    opt.DNS,
		IPAddresses: ips,
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
	// 创建网站私钥
	p, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	certDer, err := x509.CreateCertificate(rand.Reader, certCsr, rootCsr, &p.PublicKey, w.priKey)
	// certDer, err := x509.CreateCertificate(rand.Reader, certCsr, rootCsr, w.pubKey, w.priKey)
	if err != nil {
		return err
	}
	// 保存网站证书
	txt = pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDer,
	})
	err = os.WriteFile(filepath.Join(opt.OutPut, "cert.ec.pem"), txt, 0o664)
	if err != nil {
		return err
	}
	// 保存网站私钥
	txt, err = x509.MarshalECPrivateKey(p)
	if err != nil {
		return err
	}
	txt = pem.EncodeToMemory(&pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: txt,
	})
	err = os.WriteFile(filepath.Join(opt.OutPut, "cert-key.ec.pem"), txt, 0o664)
	if err != nil {
		return err
	}
	// 保存root私钥
	if opt.RootKey == "" {
		// 保存根证书
		txt = pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: rootDer,
		})
		err = os.WriteFile(filepath.Join(opt.OutPut, "root.ec.pem"), txt, 0o664)
		if err != nil {
			return err
		}
		w.ToFile("", filepath.Join(opt.OutPut, "root-key.ec.pem"))
	}
	return nil
}

// NewECC 创建一个新的ecc算法器
//
//	签名算法采用sha256
//	支持 openssl ecparam -name prime256v1/secp384r1 格式的密钥
func NewECC() *ECC {
	w := &ECC{
		signHash: NewHash(HashSHA256, nil),
	}
	return w
}
