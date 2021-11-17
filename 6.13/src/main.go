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

package main

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
	"week_4/src/packages/rsaexample"
)

func main() {
	TestSoftwareWallet()
	return
}

func TestSigningAndVerification() {
	/* Generate pseudo-random k (bit-length of the key)*/
	k := rsaexample.GenerateRandomK()
	e := 3

	/* Generate public and private key, respectively */
	publicKey, privateKey := rsaexample.KeyGen(k, e)

	/* Generate random message */
	message, _ := rand.Int(rand.Reader, publicKey.N)

	/* Hash message with SHA-256 and get integer representation of hash */
	hashedMessage := rsaexample.ByteArrayToInt(rsaexample.HashMessage(message.Bytes()))

	/* Generate RSA signature */
	signature := rsaexample.GenerateSignature(hashedMessage, publicKey)

	/* Verify RSA signature */
	rsaexample.VerifySignature(hashedMessage, signature, privateKey)

	/* Modify data, verification should fail */
	fmt.Println("Modify message.")
	message, _ = rand.Int(rand.Reader, publicKey.N)
	hashedMessage = rsaexample.ByteArrayToInt(rsaexample.HashMessage(message.Bytes()))
	rsaexample.VerifySignature(hashedMessage, signature, privateKey)
}

func MeasureHashingSpeed() {
	/* Generate 10.24KB of data */
	data := make([]byte, 10*1024)
	rand.Read(data)

	/* Measure hashing speed in bits per second */
	start := time.Now()
	rsaexample.HashMessage(data)
	time.Sleep(time.Nanosecond) //this is here because otherwise error reading the time ellapsed
	elapsed := time.Since(start)

	fmt.Printf("Hashed data (%vB) in %s \n", len(data), elapsed)
	MeasureSpeed(data, elapsed)
}

func MeasureSigningHashSpeed() {
	k := rsaexample.GenerateRandomK()
	e := 3
	publicKey, _ := rsaexample.KeyGen(k, e)
	data, _ := rand.Int(rand.Reader, publicKey.N)
	hashedMessage := rsaexample.ByteArrayToInt(rsaexample.HashMessage(data.Bytes()))
	hashedMessageSize := len(hashedMessage.Bytes())

	/* Measure hashing and signing speed in bits per second */
	start := time.Now()
	rsaexample.GenerateSignature(hashedMessage, publicKey)
	time.Sleep(time.Nanosecond)
	elapsed := time.Since(start)

	fmt.Printf("RSA key: %v bits\n", len(publicKey.N.Bytes())*8)
	fmt.Printf("Signed message hash (%vB) in %s \n", hashedMessageSize, elapsed)
	MeasureSpeed(hashedMessage.Bytes(), elapsed)
}

func MeasureSigningMessageSpeed() {
	k := rsaexample.GenerateRandomK()
	e := 3
	publicKey, _ := rsaexample.KeyGen(k, e)
	data, _ := rand.Int(rand.Reader, publicKey.N)
	dataSize := len(data.Bytes())

	/* Measure hashing and signing speed in bits per second */
	start := time.Now()
	rsaexample.GenerateSignature(data, publicKey)
	time.Sleep(time.Nanosecond)
	elapsed := time.Since(start)

	fmt.Printf("RSA key: %v bits\n", len(publicKey.N.Bytes())*8)
	fmt.Printf("Signed message (%vB) in %s \n", dataSize, elapsed)
	MeasureSpeed(data.Bytes(), elapsed)
}

func MeasureSpeed(data []byte, elapsed time.Duration) {
	/* Convert bytes per nanosecond to bits per second*/
	bytesPerNanosecond := float64(len(data)) / float64(elapsed.Nanoseconds())
	bitsPerSecond := bytesPerNanosecond * 8 * 1e9
	fmt.Printf("Speed: %v bps\n", int(bitsPerSecond))
}

func TestSoftwareWallet() {
	filename := "very secret filename"
	password := "jqrhadgoyyodsrddqtio"
	wrongPassword := "zpibmasocxjflfbfvccl"
	msg := new(big.Int).SetInt64(80085)
	publicKey := rsaexample.Generate(filename, password)

	/* Successfuly create signature */
	signature := rsaexample.Sign(filename, password, msg.Bytes())
	recoveredMsg := rsaexample.Decrypt(signature, publicKey)
	fmt.Println("Recovered message: " + recoveredMsg.String())

	/* Create signature unsuccessfully (wrong password, file with private key is not found) */
	signature = rsaexample.Sign(filename, wrongPassword, msg.Bytes())
}
