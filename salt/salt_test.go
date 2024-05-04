package salt

import (
	"crypto/aes"
	"encoding/hex"
	"encoding/json"
	"github.com/allape/stepin/env"
	"github.com/allape/stepin/stepin/create"
	"log"
	"os"
	"strings"
	"testing"
)

func TestSha256ToHexStringFromString(t *testing.T) {
	testSet := map[string]string{
		"hello":  "2CF24DBA5FB0A30E26E83B2AC5B9E29E1B161E5C1FA7425E73043362938B9824",
		"stepin": "1074618B22AA9B9C0811A1CE287D5459D7F93926CCF3E17EB6F7427A85621019",
	}
	for input, expected := range testSet {
		if Sha256ToHexStringFromString(input) != HexString(expected) {
			t.Errorf("Sha256ToHexStringFromString(%s) != %s", input, expected)
		}
	}
}

func TestNewPassword(t *testing.T) {
	if os.Getenv(string(env.StepinDatabaseSalt)) != "" || os.Getenv(string(env.StepinDatabasePassword)) != "" {
		t.Fatal("env.StepinDatabaseSalt or env.StepinDatabasePassword is not empty, test result may be affected")
	}
	extraSaltSet := map[string]string{
		"123456":   "CE4A29E3B87768E0C2B410246C620621F9BD51D5FD8C21B15863514165BEE1AA",
		"password": "1ACAF330FD909046726E8CAE60E547707D6053CAFEFAD77D6685F010D0F33756",
		"passw0rd": "FF5CFD0DF880B71BE631CEB3854195BEC6B15470067CD618F4359CD8B18F66C0",
	}
	for input, expected := range extraSaltSet {
		if strings.ToUpper(hex.EncodeToString(newPassword([]byte(input)))) != expected {
			t.Errorf("newPassword(%s) != %s", input, expected)
		}
	}
}

// text256bit
// half of a hexed sha256 string is the text password, for 256-bit AES, easy to test
func text256bit(src string) string {
	hexed := Sha256ToHexStringFromString(src)
	return string(hexed[:len(hexed)/2])
}

func TestEncode(t *testing.T) {
	plain := "hello"
	password := text256bit("world")
	iv := text256bit("iv")

	log.Println("plain:", plain)
	log.Println("password:", password)
	log.Println("iv:", iv)

	hexedIV := hex.EncodeToString([]byte(iv))
	log.Println("hexed iv:", hexedIV)

	encoded, err := encode([]byte(plain), []byte(password), []byte(iv), true)
	if err != nil {
		t.Fatal(err)
	}
	hexedEncoded := hex.EncodeToString(encoded)
	log.Println("encoded:", hexedEncoded)
	if hexedEncoded != "3041423330363832333033353636314265f967f19572379643296e12879091a7" {
		t.Fatalf("encoded != 3041423330363832333033353636314265f967f19572379643296e12879091a7")
	}

	encoded, err = encode([]byte(plain), []byte(password), []byte(iv), false)
	if err != nil {
		t.Fatal(err)
	}
	hexedEncoded = hex.EncodeToString(encoded)
	log.Println("encoded:", hexedEncoded)
	if hexedEncoded != "30414233303638323330333536363142adb9c5057192fa78fe7ca093f69c76f1" {
		t.Fatalf("encoded != 30414233303638323330333536363142adb9c5057192fa78fe7ca093f69c76f1")
	}
}

func TestDecode(t *testing.T) {
	password := text256bit("world")

	log.Println("password:", password)

	encoded, err := hex.DecodeString("3041423330363832333033353636314265f967f19572379643296e12879091a7")
	if err != nil {
		t.Fatal(err)
	}
	plain, err := decode(encoded, []byte(password), true)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("encoded:", string(plain))
	if string(plain) != "hello" {
		t.Fatalf("plain != hello")
	}

	encoded, err = hex.DecodeString("30414233303638323330333536363142adb9c5057192fa78fe7ca093f69c76f1")
	if err != nil {
		t.Fatal(err)
	}
	plain, err = decode(encoded, []byte(password), false)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("encoded:", string(plain))
	if string(plain) != "hello" {
		t.Fatalf("plain != hello")
	}
}

type SaltyCert struct {
	ID         string         `json:"id"`
	Profile    create.Profile `json:"profile"`
	Name       string         `json:"name"`
	Crt        string         `json:"crt"`
	Key        string         `json:"key"`
	Inspection string         `json:"inspection"`
}

func toCBC(content string, password []byte) (string, error) {
	b, err := hex.DecodeString(content)
	if err != nil {
		return "", err
	}
	iv := b[:aes.BlockSize]
	decoded, err := decode(b, password, false)
	if err != nil {
		return "", err
	}
	encoded, err := encode(decoded, password, iv, true)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(hex.EncodeToString(encoded)), nil
}

func TestFromNonCBC2CBC(t *testing.T) {
	jsonStr := ""

	if //goland:noinspection GoBoolExpressions
	jsonStr == "" {
		return
	}

	var body SaltyCert
	err := json.Unmarshal(
		[]byte(jsonStr),
		&body,
	)
	if err != nil {
		t.Fatal(err)
	}

	crtPassword := newPassword([]byte("_crt_salt"))
	keyPassword := newPassword([]byte("_key_salt"))
	txtPassword := newPassword([]byte("_txt_salt"))

	body.Name, err = toCBC(body.Name, txtPassword)
	if err != nil {
		t.Fatal(err)
	}
	body.Crt, err = toCBC(body.Crt, crtPassword)
	if err != nil {
		t.Fatal(err)
	}
	body.Key, err = toCBC(body.Key, keyPassword)
	if err != nil {
		t.Fatal(err)
	}
	body.Inspection, err = toCBC(body.Inspection, txtPassword)
	if err != nil {
		t.Fatal(err)
	}

	newJSON, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	log.Println(string(newJSON))
}
