package main

import (
	"fmt"
	// "encoding/binary"
	"encoding/base64"
	"encoding/hex"
	"crypto/aes"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"crypto/cipher"
	// "crypto/ed25519"
	"math/big"
	"bytes"
	"hash"
	"math"
	"io"
	
	"golang.org/x/crypto/hkdf"
	// "golang.org/x/crypto/curve25519"

	"filippo.io/edwards25519"
	"filippo.io/edwards25519/field"
)

func main() {
	/* code below used as refrence dont uncomment
	secret := []byte{0x00, 0x01, 0x02, 0x03} // this should be the root key?
	salt := []byte("salt") // this should be some shared constant?
	info := []byte("hskdf example") // this should be the dh output?

	Generate three 128-bit derived keys.
	keys := hkdf_output(2, 32, sha256.New, secret, nil, nil)
	fmt.Println(keys)
	fmt.Println()
	
	for i := range keys {
		fmt.Println(keys[i]) //byte array
		fmt.Println(b16(keys[i])) //hex representation
		fmt.Println(b64(keys[i])) //base64 representation
	}
	*/

	alice := Person{name: "alice"}
	alice.key_gen()
	bob := Person{name: "bob"}
	bob.key_gen()

	alice.X3DH_send(bob)
	bob.X3DH_recv(alice)

	// print_X3DH_diff(alice, bob)

	alice.root.state = alice.SK
	alice.root_ratchet_send()
	
	bob.root.state = bob.SK
	bob.root_ratchet_recv()

	// fmt.Println()
	// print_ratchet_diff(alice, bob)
	
	// alice.send.chain_ratchet_next(alice.send.state)
	// bob.recv.chain_ratchet_next(bob.recv.state)

	// fmt.Println()
	// print_ratchet_diff(alice, bob)

//working off of the SK from X3DH
//alice send
	alice_ciphertext1 := alice.sending("hello")
	alice.AD = nil
	//bob receive
	bob_plaintext1 := bob.recving(alice_ciphertext1)
	bob.AD = nil

//real dh ratchet tick
//bob send
	bob.otherDH = alice.myDH.PublicKey()
	bob.dh_ratchet_send()
	bob_ciphertext1 := bob.sending("world")
//alice receive
	alice.dh_ratchet_recv(bob.myDH.PublicKey())
	alice_plaintext1 := alice.recving(bob_ciphertext1)

//same dh ratchet, no tick
//bob send
	bob_ciphertext2 := bob.sending("marco")
//alice receives
	alice_plaintext2 := alice.recving(bob_ciphertext2)

//alice third
	alice.otherDH = bob.myDH.PublicKey()
	alice.dh_ratchet_send()
	alice_ciphertext2 := alice.sending("polo")

	bob.dh_ratchet_recv(alice.myDH.PublicKey())
	bob_plaintext2 := bob.recving(alice_ciphertext2)

	fmt.Println("Alice:\t", bob_plaintext1)
	fmt.Println("Bob:\t", alice_plaintext1)
	fmt.Println("Bob:\t", alice_plaintext2)
	fmt.Println("Alice:\t", bob_plaintext2)
}

func print_X3DH_diff(alice, bob Person){
	fmt.Println(alice.name)
	fmt.Println(alice.IK)
	fmt.Println(alice.EK)
	fmt.Println(alice.ScK)
	fmt.Println(alice.OPK)
	fmt.Println(alice.Xdh1)
	fmt.Println(alice.Xdh2)
	fmt.Println(alice.Xdh3)
	fmt.Println(alice.Xdh4)
	fmt.Println(alice.SK)

	fmt.Println()
	
	fmt.Println(bob.name)
	fmt.Println(bob.IK)
	fmt.Println(bob.EK)
	fmt.Println(bob.ScK)
	fmt.Println(bob.OPK)
	fmt.Println(bob.Xdh1)
	fmt.Println(bob.Xdh2)
	fmt.Println(bob.Xdh3)
	fmt.Println(bob.Xdh4)
	fmt.Println(bob.SK)
	
	fmt.Println()

	fmt.Println("name:", alice.name == bob.name)
	fmt.Println("IK:", alice.IK.Equal(bob.IK))
	fmt.Println("EK:", alice.EK.Equal(bob.EK))
	fmt.Println("ScK:", alice.ScK.Equal(bob.ScK))
	fmt.Println("OPK:", alice.OPK.Equal(bob.OPK))
	fmt.Println("Xdh1:", bytes.Equal(alice.Xdh1, bob.Xdh1))
	fmt.Println("Xdh2:", bytes.Equal(alice.Xdh2, bob.Xdh2))
	fmt.Println("Xdh3:", bytes.Equal(alice.Xdh3, bob.Xdh3))
	fmt.Println("Xdh4:", bytes.Equal(alice.Xdh4, bob.Xdh4))
	fmt.Println("len:", len(alice.SK) == len(bob.SK))
	fmt.Println("SK:", bytes.Equal(alice.SK, bob.SK))
}

func print_ratchet_diff(alice, bob Person){
	fmt.Println(alice.name)
	fmt.Println(alice.root.state)
	fmt.Println(alice.send.state)
	fmt.Println(alice.send.next)
	fmt.Println(alice.recv.state)
	fmt.Println(alice.recv.next)
	
	fmt.Println()
	
	fmt.Println(bob.name)
	fmt.Println(bob.root.state)
	fmt.Println(bob.recv.state)
	fmt.Println(bob.recv.next)
	fmt.Println(bob.send.state)
	fmt.Println(bob.send.next)

	fmt.Println()

	fmt.Println("name:", alice.name == bob.name)
	fmt.Println("root:", bytes.Equal(alice.root.state, bob.root.state))
	fmt.Println("alice-send-chain -> bob-recv-chain:", bytes.Equal(alice.send.state, bob.recv.state))
	fmt.Println("alice-send-message -> bob-recv-message:", bytes.Equal(alice.send.next, bob.recv.next))
	fmt.Println("alice-recv-chain -> bob-send-chain:", bytes.Equal(alice.recv.state, bob.send.state))
	fmt.Println("alice-recv-message -> bob-send-message:", bytes.Equal(alice.recv.next, bob.send.next))
}

var keySize = 32
var one = (&field.Element{}).One()
var pF = math.Pow(2, 255) - 19
var p = new(big.Int)
var _, _ = new(big.Float).SetFloat64(pF).Int(p)
type Ratchet struct{
	state []byte
	next []byte
}
type Person struct{
	name string

	IK, EK, ScK, OPK *ecdh.PrivateKey
	schnorrProof [][]byte
	Xdh1, Xdh2, Xdh3, Xdh4 []byte
	x, SK, AD []byte

	root Ratchet
	send Ratchet
	recv Ratchet

	myDH *ecdh.PrivateKey
	otherDH *ecdh.PublicKey
}


func (person *Person) key_gen(){
	curve := ecdh.X25519()
	person.IK, _ = curve.GenerateKey(rand.Reader)
	person.EK, _ = curve.GenerateKey(rand.Reader)
	person.ScK, _ = curve.GenerateKey(rand.Reader)
	person.OPK, _ = curve.GenerateKey(rand.Reader)
	person.myDH, _ = curve.GenerateKey(rand.Reader)



	/**
		c := sha256.Sum256(append(append(person.ScK.PublicKey().Bytes(), person.IK.PublicKey().Bytes()...), curve25519.Basepoint...))
		// cNum := new(big.Int).SetBytes(c[:])
		// cNum.Mod(cNum, p)
		// xMul, err := curve25519.X25519(c[:], person.IK.Bytes())
		// if err != nil {
			// 	fmt.Println("err: ", err)
			// }
			
		xMul := new(big.Int).SetBytes(c[:])
		xMul.Mod(xMul, p)
		ikNum := new(big.Int).SetBytes(person.IK.Bytes()[:])
		ikNum.Mod(ikNum, p)
		xMul.Mul(xMul, ikNum)
		xMul.Mod(xMul, p)
		// xNum := new(big.Int).SetBytes(xMul[:])
		scKNum := new(big.Int).SetBytes(person.ScK.Bytes()[:])
		scKNum.Mod(scKNum, p)
		xMul.Add(xMul, scKNum)
		xMul.Mod(xMul, p)

		person.x = xMul.Bytes()
		for len(person.x) < 32 {
			person.x = append([]byte{0}, person.x...)
		}
		fmt.Println(len(person.x))
		fmt.Println(person.x)
		fmt.Println()
		// 9,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0
		// person.schnorrProof = append(person.schnorrProof, curve25519.Basepoint, person.IK.PublicKey().Bytes(), person.ScK.PublicKey().Bytes(), person.x)
		// 148,82,156,147,219,75,41,147,63,124,134,21,102,138,155,219,26,12,30,26,86,68,83,138,46,215,40,248,210,34,244

		cIK, err := curve25519.X25519(c[:], person.IK.PublicKey().Bytes())
		if err != nil {
			fmt.Println("err: ", err)
		}
		fmt.Println(len(cIK))
		fmt.Println(cIK)
		fmt.Println()
		cIKNum := new(big.Int).SetBytes(cIK[:])
		cIKNum.Mod(cIKNum, p)
		scPKNum := new(big.Int).SetBytes(person.ScK.PublicKey().Bytes()[:])
		scPKNum.Mod(scPKNum, p)
		cIKNum.Add(cIKNum, scPKNum)
		cIKNum.Mod(cIKNum, p)
		cIKA := cIKNum.Bytes()
		for len(cIKA) < 32 {
			cIKA = append([]byte{0}, cIKA...)
		}
		fmt.Println(len(cIKA))
		fmt.Println(cIKA)

		xG, err := curve25519.X25519(person.x, curve25519.Basepoint)
		if err != nil {
			fmt.Println("err: ", err)
		}
		fmt.Println(len(xG))
		fmt.Println(xG)
		fmt.Println()
		fmt.Println()
	*/
}


/*
B	convert_mont(9) //from curve25519 basepoint to edwards basepoint
I	(x=0, y=1)
p	2^255 - 19
q	2^252 + 27742317777372353535851937790883648493 //ristretto255??
c	8 //group order?
d	-121665 / 121666 (mod p)
A	486662 //from curve25519 formula 
n	2 //?
|p|	255 //group order?
|q|	253 //group order?
b	256 //max byte?
*/

// xeddsa_verify(u, M, (R || s)):
// if u >= p or R.y >= 2|p| or s >= 2|q|:
// return false
// A = convert_mont(u)
// if not on_curve(A):
// return false
// h = hash(R || A || M) (mod q)
// Rcheck = sB - hA
// if bytes_equal(R, Rcheck):
// return true/false


// xeddsa_sign(k, M, Z):
// A, a = calculate_key_pair(k)
// r = hash1(a || M || Z) (mod q)
// R = rB
// h = hash(R || A || M) (mod q)
// s = r + ha (mod q)
// return R || s


// calculate_key_pair(k):
// E = kB
// A.y = E.y
// A.s = 0
// if E.s == 1:
// 	a = -k (mod q)
// else:
// 	a = k (mod q)
// return A, a
func calculate_key_pair(priv *ecdh.PrivateKey) (*edwards25519.Point, *edwards25519.Scalar) {
	var B edwards25519.Point
	var a edwards25519.Scalar

	k, err := (&edwards25519.Scalar{}).SetBytesWithClamping(priv.Bytes())
	if err != nil {
		return nil, nil
	}

	E := B.ScalarBaseMult(k).Bytes()

	if ((E[31] & 0x80) >> 7) == 1 {
		a.Negate(k)
		// Set sig bit to 0
		E[31] &= 0x7F
	} else {
		a.Set(k)
	}

	A, err := (&edwards25519.Point{}).SetBytes(E)
	if err != nil {
		return nil, nil
	}
	return A, &a
}

// convert_mont(u):
// umasked = u (mod 2|p|)
// P.y = u_to_y(umasked)
// P.s = 0
// return P
func convert_mont(u *ecdh.PublicKey) *edwards25519.Point {
	umasked, err := (&field.Element{}).SetBytes(u.Bytes())
	if err != nil {
		return nil
	}

	// y = (u - 1)/(u + 1)
	y := u_to_y(umasked).Bytes()

	// Set sign to 0
	y[31] &= 0x7F

	P, err := (&edwards25519.Point{}).SetBytes(y)
	if err != nil {
		return nil
	}

	return P
}

// The twisted Edwards curve equation is -x2 + y2 = 1 + dx2y2. 
// The u_to_y function implements the birational map 
// by calculating y = ( u - 1) * inv( u + 1) (mod p).
func u_to_y(u *field.Element) *field.Element {
	a := new(field.Element).Subtract(u, one)
	b := new(field.Element).Add(u, one)
	y := new(field.Element).Multiply(a, b.Invert(b))
	return y
}


func (person *Person) X3DH_send(otherPerson Person){
	person.Xdh1, _ = person.IK.ECDH(otherPerson.ScK.PublicKey())
	person.Xdh2, _ = person.EK.ECDH(otherPerson.IK.PublicKey())
	person.Xdh3, _ = person.EK.ECDH(otherPerson.ScK.PublicKey())
	person.Xdh4, _ = person.EK.ECDH(otherPerson.OPK.PublicKey())
	person.SK = hkdf_output(keySize, sha256.New, append(append(append(person.Xdh1, person.Xdh2...), person.Xdh3...), person.Xdh4...), nil, nil)
	person.AD = append(person.IK.PublicKey().Bytes(), otherPerson.IK.PublicKey().Bytes()...)
}

func (person *Person) X3DH_recv(otherPerson Person){
	person.Xdh1, _ = person.ScK.ECDH(otherPerson.IK.PublicKey())
	person.Xdh3, _ = person.ScK.ECDH(otherPerson.EK.PublicKey())
	person.Xdh2, _ = person.IK.ECDH(otherPerson.EK.PublicKey())
	person.Xdh4, _ = person.OPK.ECDH(otherPerson.EK.PublicKey())
	person.SK = hkdf_output(keySize, sha256.New, append(append(append(person.Xdh1, person.Xdh2...), person.Xdh3...), person.Xdh4...), nil, nil)
	person.AD = append(otherPerson.IK.PublicKey().Bytes(), person.IK.PublicKey().Bytes()...)
}


func (person *Person) root_ratchet_send(){
	person.root.root_ratchet_next(person.root.state)
	person.send.state = person.root.next
}

func (person *Person) root_ratchet_recv(){
	person.root.root_ratchet_next(person.root.state)
	person.recv.state = person.root.next
}

func (person *Person) dh_ratchet_send(){
	curve := ecdh.X25519()
	person.myDH, _ = curve.GenerateKey(rand.Reader)
	dhs, _ := person.myDH.ECDH(person.otherDH)
	person.root.root_ratchet_next(append(person.root.state, dhs...))
	person.send.state = person.root.next
}

func (person *Person) dh_ratchet_recv(dhKey *ecdh.PublicKey){
	person.otherDH = dhKey
	dhs, _ := person.myDH.ECDH(person.otherDH)
	person.root.root_ratchet_next(append(person.root.state, dhs...))
	person.recv.state = person.root.next
}

func (ratchet *Ratchet) root_ratchet_next(secret []byte) {
	output := hkdf_output(keySize*2, sha256.New, secret, nil, nil)
	ratchet.state = output[:keySize]
	ratchet.next = output[keySize:]
}

func (person *Person) sending(msg string) []byte {
	person.send.chain_ratchet_next(person.send.state)
	return encryptGCM(person.send.next, []byte(msg), person.AD)
}

func (person *Person) recving(cipherText []byte) string {
	person.recv.chain_ratchet_next(person.recv.state)
	return decryptGCM(person.recv.next, cipherText, person.AD)

}

func (ratchet *Ratchet) chain_ratchet_next(secret []byte) {
	output := hkdf_output(keySize*2, sha256.New, secret, nil, nil)
	ratchet.state = output[:keySize]
	ratchet.next = output[keySize:]
}

func hkdf_output (/*numKeys,*/ outputSize int, hash func() hash.Hash, secret, salt, info []byte) []byte{
	hkdf_step := hkdf.New(hash, secret, salt, info)

	var keys []byte
	key := make([]byte, outputSize)
	if _, err := io.ReadFull(hkdf_step, key); err != nil {
		panic(err)
	}

	keys = key
	return keys
}


func encryptGCM(key, msg, AD []byte) []byte {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil
	}

	cipherBytes := gcm.Seal(nonce, nonce, msg, AD)

	return cipherBytes
}

func decryptGCM(key, cipherBytes, AD []byte) string {

	block, err := aes.NewCipher(key)
	if err != nil {
		return "err"
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "err"
	}
	if len(cipherBytes) < gcm.NonceSize() {
		return "err"
	}

	nonce := cipherBytes[:gcm.NonceSize()]
	cipherBytes = cipherBytes[gcm.NonceSize():]

	plainBytes, err := gcm.Open(nil, nonce, cipherBytes, AD)
	if err != nil {
		return "err"
	}

	return string(plainBytes)
}


func b16(msg []byte) string {
	return hex.EncodeToString(msg)
}
func b64(msg []byte) string {
	return base64.StdEncoding.EncodeToString(msg)
}

