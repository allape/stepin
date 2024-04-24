package salt

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/allape/stepin/env"
	"strings"
)

type HexString string

var (
	Salt     = env.Get(env.StepinDatabaseSalt, "stepin is salty, like ocean-favor potato chips")
	Password = env.Get(env.StepinDatabasePassword, "123456")
)

func Sha256(content []byte) []byte {
	hash := sha256.New()
	hash.Write(content)
	return hash.Sum(nil)
}

func Sha256ToHexString(content []byte) HexString {
	return HexString(strings.ToUpper(hex.EncodeToString(Sha256(content))))
}

func Sha256ToHexStringFromString(content string) HexString {
	return HexString(strings.ToUpper(hex.EncodeToString(Sha256([]byte(content)))))
}

func Encode(plain, salt []byte) ([]byte, error) {
	iv := make([]byte, aes.BlockSize)
	_, err := rand.Read(iv)
	if err != nil {
		return nil, err
	}

	password := Sha256(append([]byte(Salt+Password), salt...))

	paddingLength := aes.BlockSize - (len(plain) % aes.BlockSize)
	plain = append(plain, bytes.Repeat([]byte{byte(paddingLength)}, paddingLength)...)

	cipher, err := aes.NewCipher(password)
	if err != nil {
		return nil, err
	}

	content := make([]byte, len(plain))
	cipher.Encrypt(content, iv)
	for index := range len(plain) / aes.BlockSize {
		i1 := index * aes.BlockSize
		i2 := (index + 1) * aes.BlockSize
		cipher.Encrypt(content[i1:i2], plain[i1:i2])
	}

	return append(iv, content...), nil
}

func Decode(content, salt []byte) ([]byte, error) {
	if len(content)%aes.BlockSize != 0 || len(content) < aes.BlockSize*2 {
		return nil, errors.New("invalid content length")
	}

	password := Sha256(append([]byte(Salt+Password), salt...))

	iv := content[:aes.BlockSize]
	content = content[aes.BlockSize:]

	cipher, err := aes.NewCipher(password)
	if err != nil {
		return nil, err
	}

	plain := make([]byte, len(content))
	cipher.Decrypt(plain, iv)
	for index := range len(content) / aes.BlockSize {
		i1 := index * aes.BlockSize
		i2 := (index + 1) * aes.BlockSize
		cipher.Decrypt(plain[i1:i2], content[i1:i2])
	}

	paddingSize := plain[len(plain)-1]

	return plain[:len(plain)-int(paddingSize)], nil
}

func EncodeToHexString(plain, salt []byte) (HexString, error) {
	content, err := Encode(plain, salt)
	if err != nil {
		return "", err
	}
	return HexString(strings.ToUpper(hex.EncodeToString(content))), nil
}

func DecodeFromHexString(hexStr HexString, salt []byte) ([]byte, error) {
	content, err := hex.DecodeString(string(hexStr))
	if err != nil {
		return nil, err
	}
	plain, err := Decode(content, salt)
	if err != nil {
		return nil, err
	}
	return plain, nil
}

func DecodeFromHexStringToString(hexStr HexString, salt []byte) (string, error) {
	plain, err := DecodeFromHexString(hexStr, salt)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
