/**
BY: Deyana Atanasova, Henrik Tambo Buhl & Alexander Stæhr Johansen
DATE: 31-10-2021
COURSE: Distributed Systems and Security
DESCRIPTION: RSA en- and decryption template implementation.
**/

/**
The implementation is based on the book "Secure Distributed Systems" 2021,
section 5.2.1 by Ivan Damgaard, Jesper Buus Nielsen & Claudio Orlandi.
**/

package rsaexample

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
)

/* Key struct */
type Key struct {
	N      *big.Int
	E_or_d *big.Int
}

/* Get a byte array representation of Key */
func (key Key) ToBytes() []byte {
	keyBytes, err := json.Marshal(key)
	if err != nil {
		panic(err)
	}
	return keyBytes
}

/* Decode bytes to key */
func ToKey(keyBytes []byte) Key {
	var key Key
	json.Unmarshal(keyBytes, &key)
	return key
}

/* Key generator method */
func KeyGen(K *big.Int, e int) (Key, Key) {
	/* Convert constants 1, E and K to big ints */
	ONE := big.NewInt(1)
	E := big.NewInt(int64(e))

	/* Determine bitlength of k */
	bitLength := K.BitLen()

	/* Step 1: Generate prime, p, with half the bitlength of k.
	   - The reason is that the product of two numbers with bitlength n/2 is n */
	p, _ := rand.Prime(rand.Reader, bitLength/2)

	/* Step 2: Subtract 1 from p */
	P := new(big.Int).Sub(p, ONE)

	/* Step 3: Find GCD between E and P */
	gcd_1 := new(big.Int).GCD(nil, nil, E, P)

	/* For GCD != 1, repeat steps 1, 2 and 3 */
	for ONE.Cmp(gcd_1) != 0 {
		p, _ = rand.Prime(rand.Reader, bitLength/2)
		P = new(big.Int).Sub(p, ONE)
		gcd_1 = new(big.Int).GCD(nil, nil, E, P)
	}

	/* Generate prime q applying same procedure as explained for p */
	q, _ := rand.Prime(rand.Reader, bitLength/2)
	Q := new(big.Int).Sub(q, ONE)
	gcd_2 := new(big.Int).GCD(nil, nil, E, Q)
	for ONE.Cmp(gcd_2) != 0 && p.Cmp(q) != 0 {
		q, _ = rand.Prime(rand.Reader, bitLength/2)
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

/* Hash message */
func HashMessage(m []byte) []byte {
	hm := sha256.Sum256(m)
	return hm[:]
}

/* Turn a byte array into an integer */
func ByteArrayToInt(inputBytes []byte) *big.Int {
	return new(big.Int).SetBytes(inputBytes[:])
}

/* Generate RSA signature */
func GenerateSignature(hashedMessage *big.Int, publicKey Key) *big.Int {
	/* Encrypt the hashed message with the public key */
	ciphertext := Encrypt(hashedMessage, publicKey)

	/* Pad ciphertext with zeros */
	ciphertextInBytes := ciphertext.Bytes()
	keyInBytes := publicKey.N.Bytes()
	if len(ciphertextInBytes) < len(keyInBytes) {
		padding := make([]byte, len(keyInBytes)-len(ciphertextInBytes))
		ciphertextInBytes = append(padding, ciphertextInBytes...)
	}

	return new(big.Int).SetBytes(ciphertextInBytes)
}

/* Verify signature */
func VerifySignature(hashedMessage *big.Int, ciphertext *big.Int, privateKey Key) {
	/* Decrypt signature */
	decryptedHashedMessage := Decrypt(ciphertext, privateKey)

	/* Compare the hashed message and the hash of the message from the signature */
	if hashedMessage.Cmp(decryptedHashedMessage) == 0 {
		fmt.Println("Message hash and decrypted message hash match.")
	} else {
		fmt.Println("Message hash and decrypted message hash do not match.")
	}
}

/* Generate pseudo-random k (bit-length of the key)*/
func GenerateRandomK() *big.Int {
	max := new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(2048), nil).Sub(max, big.NewInt(1))
	k, err := rand.Int(rand.Reader, max)
	if err != nil {
		fmt.Println(err)
	}
	return k
}

func Generate(filename string, password string) Key {
	/* Encrypt filename with the password using an XOR cipher */
	if len(filename) != len(password) {
		fmt.Println("Filename and password must be the same size.")
	}
	filenameBytes := []byte(filename)
	passwordBytes := []byte(password)
	encryptedFilenameBytes := make([]byte, len(filenameBytes))
	for i := 0; i < len(passwordBytes); i++ {
		encryptedFilenameBytes[i] = filenameBytes[i] ^ passwordBytes[i]
	}
	encryptedFilenameHex := hex.EncodeToString(encryptedFilenameBytes)

	/* Generate public and private key */
	k := GenerateRandomK()
	e := 3
	publicKey, privateKey := KeyGen(k, e)

	/* Save private key in a file with the encrypted filename */
	err := ioutil.WriteFile(encryptedFilenameHex+".data", privateKey.ToBytes(), 0777)
	if err != nil {
		fmt.Println(err)
	}

	/* Return public key */
	return publicKey
}

func Sign(filename string, password string, msg []byte) *big.Int {
	/* Encrypt filename with the password using an XOR cipher */
	if len(filename) != len(password) {
		fmt.Println("Filename and password must be the same size.")
	}
	filenameBytes := []byte(filename)
	passwordBytes := []byte(password)
	encryptedFilenameBytes := make([]byte, len(filenameBytes))
	for i := 0; i < len(passwordBytes); i++ {
		encryptedFilenameBytes[i] = filenameBytes[i] ^ passwordBytes[i]
	}
	encryptedFilenameHex := hex.EncodeToString(encryptedFilenameBytes)

	/* If file with the encrypted filename exists, read private key*/
	privateKeyBytes, err := os.ReadFile(encryptedFilenameHex + ".data")
	if err != nil {
		log.Fatal("Wrong filename/password.")
	}
	privateKey := ToKey(privateKeyBytes)

	/* Sign message */
	msgInt := ByteArrayToInt(msg)
	signature := Encrypt(msgInt, privateKey)

	/* Return signature */
	return signature
}
