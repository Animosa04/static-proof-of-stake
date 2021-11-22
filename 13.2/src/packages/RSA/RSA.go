package RSA

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"hash"
	"math/big"
	mrand "math/rand"
	"time"
)

/* Key struct */
type Key struct {
	N      *big.Int
	E_or_d *big.Int
}

/* Encode key to string */
func (key *Key) ToString() string {
	keyString, err := json.Marshal(key)
	if err != nil {
		panic(err)
	}
	return string(keyString)
}

/* Decode string to key */
func ToKey(keyString string) Key {
	var key Key
	json.Unmarshal([]byte(keyString), &key)
	return key
}

/* Generate pseudo-random k (bit-length of the key)*/
func GenerateRandomK() int {
	s1 := mrand.NewSource(time.Now().UnixNano())
	r1 := mrand.New(s1)
	min := 512
	max := 1024
	return r1.Intn(max-min) + min // Factory grade keys. Adjust max val for faster computation...
}

/* Key generator method */
func KeyGen(K int, e int) (Key, Key) {
	/* Convert constants 1, E and K to big ints */
	ONE := big.NewInt(1)
	E := big.NewInt(int64(e))

	/* Step 1: Generate prime, p, with half the bitlength of k.
	   - The reason is that the product of two numbers with bitlength n/2 is n */
	p, _ := rand.Prime(rand.Reader, K/2)

	/* Step 2: Subtract 1 from p */
	P := new(big.Int).Sub(p, ONE)

	/* Step 3: Find GCD between E and P */
	gcd_1 := new(big.Int).GCD(nil, nil, E, P)

	/* For GCD != 1, repeat steps 1, 2 and 3 */
	for ONE.Cmp(gcd_1) != 0 {
		p, _ = rand.Prime(rand.Reader, K/2)
		P = new(big.Int).Sub(p, ONE)
		gcd_1 = new(big.Int).GCD(nil, nil, E, P)
	}

	/* Generate prime q applying same procedure as explained for p */
	q, _ := rand.Prime(rand.Reader, K/2)
	Q := new(big.Int).Sub(q, ONE)
	gcd_2 := new(big.Int).GCD(nil, nil, E, Q)
	for ONE.Cmp(gcd_2) != 0 && p.Cmp(q) != 0 {
		q, _ = rand.Prime(rand.Reader, K/2)
		Q = new(big.Int).Sub(q, ONE)
		gcd_2 = new(big.Int).GCD(nil, nil, E, Q)
	}

	/* Generate public key as (n, e) */
	publicKey := Key{N: new(big.Int).Mul(p, q), E_or_d: E}

	/* Generate private key as (n, d) */
	privateKey := Key{N: new(big.Int).Mul(p, q), E_or_d: new(big.Int).ModInverse(E, new(big.Int).Mul(P, Q))}
	return publicKey, privateKey
}

/* Encrypt method */
func Encrypt(M *big.Int, privateKey Key) *big.Int {
	/* Generate ciphertext using the private key*/
	c := new(big.Int).Exp(M, privateKey.E_or_d, privateKey.N)
	return c
}

/* Decrypt method */
func Decrypt(c *big.Int, publicKey Key) *big.Int {
	/* Decrypt the message using the public key */
	m := new(big.Int).Exp(c, publicKey.E_or_d, publicKey.N)
	return m
}

/* Template hashing method */
// TODO: consider adding to different objects
func ComputeHash(templateObject interface{}) []byte {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", templateObject)))
	return h.Sum(nil)
}

func AddToHash(h *hash.Hash, str string) {
	if _, err := (*h).Write([]byte(str)); err != nil {
		panic(err)
	}
}

/* Turn a byte array into an integer */
func ByteArrayToInt(inputBytes []byte) *big.Int {
	return new(big.Int).SetBytes(inputBytes[:])
}

/* Generate RSA signature */
func GenerateSignature(templateObject interface{}, privateKeyString string) string {
	/* Hash object with SHA-256 and get integer representation of hash, */
	templateObjectHash := ByteArrayToInt(ComputeHash(templateObject))

	/* Turn the string-encoded private key into Key */
	privateKey := ToKey(privateKeyString)

	/* Encrypt the hashed transaction with the private key */
	ciphertext := Encrypt(templateObjectHash, privateKey)

	/* Pad ciphertext with zeros */
	ciphertextInBytes := ciphertext.Bytes()
	keyInBytes := privateKey.N.Bytes()
	if len(ciphertextInBytes) < len(keyInBytes) {
		padding := make([]byte, len(keyInBytes)-len(ciphertextInBytes))
		ciphertextInBytes = append(padding, ciphertextInBytes...)
	}
	signature := new(big.Int).SetBytes(ciphertextInBytes)
	return signature.String()
}

/* Verify signature */
func VerifySignature(templateObject interface{}, signature string, publicKeyString string) bool {
	/* Hash object with SHA-256 and get integer representation of hash, */
	templateObjectHash := ByteArrayToInt(ComputeHash(templateObject))

	/* Convert the signature to a big.Int */
	convertedSignature, _ := new(big.Int).SetString(signature, 10)

	/* Turn the string-encoded private key into Key */
	publicKey := ToKey(publicKeyString)

	/* Decrypt signature */
	decryptedHashedMessage := Decrypt(convertedSignature, publicKey)

	/* Compare the hashed message and the hash of the message from the signature */
	if templateObjectHash.Cmp(decryptedHashedMessage) == 0 {
		return true
	} else {
		return false
	}
}

// So it goes like this:
/**
	SomeObjectToSign object
	PrivateKey string

	SomeObjectToSign -> []byte (hash) -> bigInt
	PrivateKey -> Key

	Encrypt(bigInt with Key) -> bigInt (ciphertext) -> []byte (pad ciphertext) -> bigInt -> string

	GenerateSignature (object, string) returns signature in string (but could as easily return just bytes, or bigInt)

	key.ToString() to turn into string

**/
