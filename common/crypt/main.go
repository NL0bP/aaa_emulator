package crypt

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"encoding/hex"
	"encoding/pem"
	"fmt"
)

// Crc8 ... Checksum for server-side encrypted packets
func Crc8(packet []byte) byte {
	checksum := byte(0)
	for i := 0; i < len(packet); i++ {
		checksum = checksum * 19
		checksum += packet[i]
	}
	return checksum & 255
}

// toClientEncr help function
func add(key *uint) byte {
	*key += 3132373
	n := (*key >> 16) & 247
	if n == 0 {
		n = 254
	}
	return byte(n)
}

//ToClientEncr ... encrypt message to client
func ToClientEncr(packet []byte) []byte {
	length := len(packet)
	array := make([]byte, length)
	key := uint(length ^ 522286496)
	n := 4 * int(length/4)

	for i := n - 1; i >= 0; i-- {
		val := add(&key)
		array[i] = packet[i] ^ val
	}
	for i := n; i < length; i++ {
		val := add(&key)
		array[i] = packet[i] ^ val
	}
	return array
}

// CryptRSA ...
type CryptRSA struct {
	pubKey  *rsa.PublicKey
	privKey *rsa.PrivateKey
}

// LoadRSA ... loads predifined rsa keys and returns CryptRSA object
func LoadRSA() *CryptRSA {
	pemPub := "-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCjirTvOfa4UrhpApiFWitJTSGv\nSLQ4IoUk20C1q7G+Zep3PyEWtlt00RP9w/fPkaAukMuFjiDG2VTEaQe5Oczv2OjY\n5GvpYgihqqd2qCXdJheo+v4ncDI1nAWu2WuwLn0idEjoFhnI55kXhalPAzD87TnE\nD43FWgAcu5tCbL6obwIDAQAB\n-----END PUBLIC KEY-----"
	pemPriv := "-----BEGIN RSA PRIVATE KEY-----\nMIICXAIBAAKBgQCjirTvOfa4UrhpApiFWitJTSGvSLQ4IoUk20C1q7G+Zep3PyEW\ntlt00RP9w/fPkaAukMuFjiDG2VTEaQe5Oczv2OjY5GvpYgihqqd2qCXdJheo+v4n\ncDI1nAWu2WuwLn0idEjoFhnI55kXhalPAzD87TnED43FWgAcu5tCbL6obwIDAQAB\nAoGAS0KoXFUI6K9MpSqoJOondG61+zPSl+iu7BSoNVKDlBLTsTfQkuKtuNcEw6n8\n7z1dgUBqIJaVF91pCJArGT7zw4mBhSKbBMTkVPk3KlJUpGHVSNDuSO3hQ/7MuTD7\nbErf2OWAbpEq6e+BJknCp0yckc69+olRNwnZ1GiHVmHfKR0CQQC5GRUPfw0kMKmT\nq0NeWwS61dcgYGm8CQiRqeYfQ8dl0BvQAEeXPRz0eMCws8IHI5lDlpXLDWDD6IVt\n0U7POI3lAkEA4i/NtgBb2YmkHqyebj7JyAGdGz+uIGJAbmPhxRZ1oErEuCkIg27c\n9y3Al2c5aE/diJyUK5Lj0uyKeIzkeY8XwwJAVIvXadefugscOh49TGkItQqeE+TW\nBxSdPGO9gERmXOP9ADpQeQ1qH2TUpyHEm5wwEoZC75exvmqEH9A+TjrH3QJBAN5u\nhk0CQ1FFo2kq9k6SXpraw2ZllFZyaMxmW0MXWCt++7/jUmT2ZESL8Mazk2f6inBr\nEuda98KYLYBphdHpH0MCQBCyMlTdr4O/0GvG7iY12EG8WkhCrKqqpZa4CFw42Ho3\nKkGaXDNQ02ugSWTCLNJL7bPa25j57ncMZMRSSpcFh08=\n-----END RSA PRIVATE KEY-----"

	blockPub, _ := pem.Decode([]byte(pemPub))
	blockPriv, _ := pem.Decode([]byte(pemPriv))

	pubKey, _ := x509.ParsePKIXPublicKey(blockPub.Bytes)
	privKey, _ := x509.ParsePKCS1PrivateKey(blockPriv.Bytes)

	_rsa := new(CryptRSA)
	_rsa.pubKey = pubKey.(*rsa.PublicKey)
	_rsa.privKey = privKey
	return _rsa
}

// GetXorKey ... extracts XOR key
func (cr *CryptRSA) GetXorKey(raw []byte) uint {
	rng := rand.Reader
	keyXORraw, err := rsa.DecryptPKCS1v15(rng, cr.privKey, raw)
	if err != nil {
		fmt.Println("Error", err)
	}

	head := binary.LittleEndian.Uint32(keyXORraw[:4])
	keyXOR := (head^0x15a0248e)*head ^ 0x070f1f23&0xffffffff

	return uint(keyXOR)
}

// GetAesKey ... extracts AES key
func (cr *CryptRSA) GetAesKey(raw []byte) []byte {
	rng := rand.Reader
	keyAES, err := rsa.DecryptPKCS1v15(rng, cr.privKey, raw)
	if err != nil {
		fmt.Println("Error", err)
	}
	return keyAES
}

// CryptAES ... decrypt packets encrypted with AES
type CryptAES struct {
	aesKey []byte
	xorKey uint
	msgKey map[uint8]uint8
	seq    uint
	mode   cipher.BlockMode
}

// ClientCrypt ... Decrypt packets from client
func ClientCrypt(aesKey []byte, xorKey uint) *CryptAES {
	_aes := new(CryptAES)
	_aes.aesKey = aesKey
	_aes.xorKey = xorKey * xorKey & 0xffffffff
	_aes.msgKey = map[uint8]uint8{0x30: 0x2f, 0x31: 0x2, 0x33: 0x4, 0x34: 0x5, 0x35: 0x16, 0x36: 0x27, 0x37: 0x8, 0x38: 9, 0x39: 0xa, 0x3b: 0x1c, 0x3c: 0xd, 0x3e: 0x1f, 0x3f: 0x10}
	_aes.seq = 0

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		panic(err)
	}
	iv := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	_aes.mode = cipher.NewCBCDecrypter(block, iv)
	return _aes
}

func (cr *CryptAES) decXor(packet []byte, mkey uint8, offset int) []byte {
	length := len(packet)
	array := make([]byte, length)
	mul := cr.xorKey * uint(mkey)
	key := (0x75a024a4 ^ mul) ^ 0xC3903b6a

	n := 4 * int(length/4)

	for i := n - 1 - offset; i >= 0; i-- {
		val := add(&key)
		array[i] = packet[i] ^ val
	}
	for i := n - offset; i < length; i++ {
		val := add(&key)
		array[i] = packet[i] ^ val
	}
	return array
}

// Decrypt ...
func (cr *CryptAES) Decrypt(data []byte, size int) []byte {
	defer func() {
		cr.seq++
	}()
	if (len(data) - 1) < aes.BlockSize {
		panic("ciphertext too short")
	}
	// CBC mode always works in whole blocks.
	if (len(data)-1)%aes.BlockSize != 0 {
		panic("ciphertext is not a multiple of the block size")
	}

	xored := make([]byte, size)
	if _, ok := cr.msgKey[data[0]]; !ok {
		fmt.Println("[CRYPT] No key in map:", data[0])
	}
	mkey := cr.msgKey[data[0]]
	msg := data[1 : size-2]
	//fmt.Print("Mkey: ", mkey, " ")

	if cr.seq == 0 {
		xored = cr.decXor(msg, mkey, 7)
		//fmt.Print("Off: ", 7, "  ")
	} else if cr.seq == 1 {
		xored = cr.decXor(msg, mkey, 1)
		//fmt.Print("Off: ", 1, "  ")
	} else if cr.seq == 2 || cr.seq == 3 || cr.seq == 4 {
		xored = cr.decXor(msg, mkey, 0)
		//fmt.Print("Off: ", 0, "  ")
	} else {
		xored = cr.decXor(msg, mkey, 0)
		//fmt.Print("Off: ", 0, "  ")
	}
	decr := make([]byte, size)
	cr.mode.CryptBlocks(decr, xored)
	return decr
}

func main() {
	rsa := LoadRSA()

	keys, _ := hex.DecodeString("02422645d225e0a0705c8ffd3fc07153c434dce752e614bc38734b1a470b16a5007936955658c2028784d2677203165c7c2270245a4e8a5414b0171dec91c9ac18330285b0815cae7d49e3808e3103ad95ff8664feb2498798f589494832a422a64fb8eb4a6257b2c9d678f129f28d4423f62da14a2985e7f3324030c56a6f24832ef5829af741100d9eeb06cad2067fcb358012fe7cd4b869722c1fdc2af1b7d2fe2ed091b5c2cfaafb670c4edbc336b5791aff8d8dddab8005458cd6f02d5e2e134d5df810891a85d1739e9ffc3777e9f4bedb3ab432ffbcceb14e549ab091bfddc8fdb8c9bbede43a245fc0f2bdeb869af341476567dacc18f9ff910866b7")
	aesKey := rsa.GetAesKey(keys[:128])
	xorKey := rsa.GetXorKey(keys[128:])
	fmt.Println("AES:", hex.EncodeToString(aesKey), " XOR:", xorKey)
	crypt := ClientCrypt(aesKey, xorKey)

	pack1, _ := hex.DecodeString("390134260e5f08d64f4621e003daaf0068")
	dec := crypt.Decrypt(pack1, 19)
	fmt.Println(hex.EncodeToString(dec))

	pack1, _ = hex.DecodeString("39ee448d628e8bbd7273cbe605a2b73341")
	dec = crypt.Decrypt(pack1, 19)
	fmt.Println(hex.EncodeToString(dec))

	pack1, _ = hex.DecodeString("37479a519f06a3e8a7e9de9b9d6fad50e0")
	dec = crypt.Decrypt(pack1, 19)
	fmt.Println(hex.EncodeToString(dec))

	//s, _ := hex.DecodeString("86764616e6b6875727f7c7726430fed1a1714111e2b2825222f7c297623300d4a4d7cea10a8c73ed744eaf94feb25dfcee3a718fb874a843b4250ae1c7e9a35cd769241cd2d2223f40d5c658b6b2da7716a8c6ed7249b7a1eeaa14c977f9282d5e59b9fa16a97b003ba2790405ec31399293fddf0be2e554039ad309ae2ca7c9bdb2147816c7b8b9a688f5372b1d21c33f7e5af70b59612e44095e2ec739985ea996663707d7a7774010e0b0815121f1c192623202d2a3734313e3b4845424f4c595653505d6a6764616e7b7875727fed0a0704011e1b1815122f2c292623303d3a3744414e4b4855525f5c596663606d6a7774717e7b0805020f0c191613101d2a2724212e3b3835323f4c494643405d5a5754516e6b6865727f7c797704110e1ba8050a4aeea")
	//fmt.Println(hex.EncodeToString(ToClientEncr(s)))
}
