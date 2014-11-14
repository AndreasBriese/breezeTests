package main

import (
	"bytes"
	"crypto/md5"
	crand "crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"sync"
	"time"
	// "unsafe"

	// "./breeze"
	"github.com/AndreasBriese/breeze"
)

var mutex = &sync.Mutex{}

//
// MWC
//

const cmwcPHI = 0x9e3779b9

type cmwcRand struct {
	Q     [4096]uint32
	i     uint32
	carry uint32
}

func (m *cmwcRand) Init(x uint32) {
	m.carry = 362436
	m.i = 4095

	m.Q[0] = x
	m.Q[1] = x + cmwcPHI
	m.Q[2] = x + cmwcPHI + cmwcPHI

	for i := uint32(3); i < 4096; i++ {
		m.Q[i] = m.Q[i-3] ^ m.Q[i-2] ^ cmwcPHI ^ i
	}

}

func (m *cmwcRand) rand_cmwc() uint32 {

	m.i = (m.i + 1) & 4095
	t := (18705 * m.Q[m.i]) + m.carry
	m.carry = t >> 32
	x := t + m.carry

	if x < m.carry {
		x++
		m.carry++
	}

	m.Q[m.i] = 0xfffffffe - x

	return m.Q[m.i]
}

func (m *cmwcRand) rand_cmwcMP() uint32 {
	mutex.Lock()

	m.i = (m.i + 1) & 4095

	t := (18705 * m.Q[m.i]) + m.carry
	m.carry = t >> 32
	x := t + m.carry

	if x < m.carry {
		x++
		m.carry++
	}

	m.Q[m.i] = 0xfffffffe - x
	mutex.Unlock()
	return m.Q[m.i]
}

//
// salsa20
//
type prnGenSalsa struct {
	counter  [16]byte
	block    [64]byte
	keyBytes [32]byte
	idx      int8
}

func (s *prnGenSalsa) Init(keyPhrase string) {
	hash := sha512.New()
	io.WriteString(hash, keyPhrase)
	key := hash.Sum(nil)
	copy(s.keyBytes[:], key[:])
	core(&(s.block), &(s.counter), &(s.keyBytes), &Sigma)
}

func (s *prnGenSalsa) reset() {
	s.counter = [16]byte{}
	s.idx = 0
	core(&(s.block), &(s.counter), &(s.keyBytes), &Sigma)
}

func (s *prnGenSalsa) nextByte() (byt int) {
	byt = int(s.block[s.idx])
	s.idx++
	if s.idx == 64 {
		u := uint32(1)
		for i := 8; i < 16; i++ {
			u += uint32(s.counter[i])
			s.counter[i] = byte(u)
			u >>= 8
		}
		core(&(s.block), &(s.counter), &(s.keyBytes), &Sigma)
		s.idx = 0
	}
	return byt
}

func (s *prnGenSalsa) nextByteMP() (byt int) {
	mutex.Lock()
	byt = int(s.block[s.idx])
	s.idx++
	if s.idx == 64 {
		u := uint32(1)
		for i := 8; i < 16; i++ {
			u += uint32(s.counter[i])
			s.counter[i] = byte(u)
			u >>= 8
		}
		core(&(s.block), &(s.counter), &(s.keyBytes), &Sigma)
		s.idx = 0
	}
	mutex.Unlock()
	return byt
}

//
// salsa XOR
//

//
// salsa20
//
type prnGenXORSalsa struct {
	counter  [16]byte
	block    [64]byte
	keyBytes [32]byte
	idx      int8
}

func (s *prnGenXORSalsa) XOR(out, in []byte, keyPhrase string) {
	hash := sha512.New()
	io.WriteString(hash, keyPhrase)
	key := hash.Sum(nil)
	copy(s.keyBytes[:], key[:])
	// core(&(s.block), &(s.counter), &(s.keyBytes), &Sigma)
	xORKeyStream(out, in, &(s.counter), &(s.keyBytes))
}

//
// Crypto.rand
var b = make([]byte, n)

func testCryptoRand(n int) {
	// var b = make([]byte, n)
	_, err := crand.Read(b)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	// The slice should now contain random bytes instead of only zeroes.
	if bytes.Equal(b, make([]byte, n)) {
		fmt.Println("err")
	}
}

var (
	n  = 100000000 // 100 Mio
	st time.Time
)

func main() {

	fmt.Println("\nStreamlength (bytes):", n)

	fmt.Println("\n Initialization timings\n")

	rand.Seed(12345678)

	st = time.Now()
	var cmwc cmwcRand
	cmwc.Init(uint32(12345678))
	fmt.Println("cmwcRand.init", time.Since(st).Nanoseconds(), "ns/op")

	st = time.Now()
	var slsa prnGenSalsa
	slsa.Init("12345678")
	fmt.Println("salsa.init", time.Since(st).Nanoseconds(), "ns/op")

	// var lmap LogistMap
	// lmap.Seed(uint64(12345678))

	// st = time.Now()
	// var lmap2 LogistMap2
	// lmap2.Init(uint64(12345678))
	// fmt.Println("breeze64_32i.init", time.Since(st).Nanoseconds(), "ns/op")

	// st = time.Now()
	// var bmap breeze.Breeze64_32
	// bmap.Init("1234567")
	// fmt.Println("breeze64_32.init", time.Since(st).Nanoseconds(), "ns/op")

	// st = time.Now()
	// var bmap72 breeze.Breeze128_72
	// bmap72.Init("1234567")
	// fmt.Println("breeze128_72.init", time.Since(st).Nanoseconds(), "ns/op")

	st = time.Now()
	var bmap128 breeze.Breeze128
	bmap128.Init("12345678")
	fmt.Println("breeze128.init", time.Since(st).Nanoseconds(), "ns/op")

	st = time.Now()
	var bmapCS128 breeze.BreezeCS128
	bmapCS128.Init()
	fmt.Println("breezeCS128.init", time.Since(st).Nanoseconds(), "ns/op")

	st = time.Now()
	var bmap256 breeze.Breeze256
	bmap256.Init("12345678" + "12345678")
	fmt.Println("breeze256.init", time.Since(st).Nanoseconds(), "ns/op")

	st = time.Now()
	var bmap512 breeze.Breeze512
	bmap512.Init("12345678" + "12345678")
	fmt.Println("breeze512.init", time.Since(st).Nanoseconds(), "ns/op")

	var byt uint8
	var bts = make([]int, n)

	fmt.Println("\nTimings without initialisation")

	for j := 1; j < 4; j++ {
		fmt.Println("\n round", j, "\n")

		sd := uint64(time.Now().Nanosecond())

		// lmap.Seed(uint64(sd))
		// lmap2.Init(sd)
		// bmap.Reset()
		// bmap.Init(sd)
		// bmap72.Reset()
		// bmap72.Init(sd)
		bmap128.Reset()
		err := bmap128.Init(sd)
		if err != nil {
			fmt.Println(err)
			panic(1)
		}
		bmapCS128.Reset()
		// err := bmap128CS.Init(sd)
		if err != nil {
			fmt.Println(err)
			panic(1)
		}
		bmap256.Reset()
		err = bmap256.Init([]uint64{sd, sd})
		if err != nil {
			fmt.Println(err)
			panic(1)
		}
		bmap512.Reset()
		err = bmap512.Init([]uint64{sd, sd, sd, sd})
		if err != nil {
			fmt.Println(err)
			panic(1)
		}

		slsa = prnGenSalsa{}
		slsa.Init(fmt.Sprintf("%v", sd))

		st = time.Now()
		testCryptoRand(n)
		fmt.Println("crypto/rand", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		st = time.Now()
		for i := 0; i < n; i++ {
			_ = rand.Int()
		}
		fmt.Println("math/rand", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		// st = time.Now()
		// for i := 0; i < n; i++ {
		// 	lmap.Byte(&byt)
		// }
		// fmt.Println("breeze8", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		// st = time.Now()
		// for i := 0; i < n; i++ {
		// 	lmap2.Byte(&byt)
		// }
		// fmt.Println("breeze64_32i", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		// st = time.Now()
		// for i := 0; i < n; i++ {
		// 	bmap.Byte(&byt)
		// }
		// fmt.Println("breeze64_32", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		st = time.Now()
		for i := 0; i < n; i++ {
			bmap128.RandIntn()
			// if i < 1000 {
			// 	fmt.Print(rd, ",")
			// }
		}
		fmt.Println("breeze128 rdInt", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		// st = time.Now()
		// for i := 0; i < n; i++ {
		// 	bmap72.Byte(&byt)
		// 	// if i < 1000 {
		// 	// 	fmt.Print(byt, ",")
		// 	// }
		// }
		// fmt.Println("breeze128_72", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		st = time.Now()
		for i := 0; i < n; i++ {
			bmap128.RandDbl()
			// if i < 1000 {
			// 	fmt.Print(rd, ",")
			// }
		}
		fmt.Println("breeze128 rdDbl", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		st = time.Now()
		for i := 0; i < n; i++ {
			bmap128.Byte(&byt)
			// if i < 1000 {
			// 	fmt.Print(byt, ",")
			// }
		}
		fmt.Println("breeze128", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		st = time.Now()
		for i := 0; i < n; i++ {
			bmapCS128.Byte(&byt)
			// if i < 1000 {
			// 	fmt.Print(byt, ",")
			// }
		}
		fmt.Println("breezeCS128", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		st = time.Now()
		for i := 0; i < n; i++ {
			bmap256.Byte(&byt)
			// if i < 1000 {
			// 	fmt.Print(byt, ",")
			// }
		}
		fmt.Println("breeze256", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		st = time.Now()
		for i := 0; i < n; i++ {
			bmap512.Byte(&byt)
			// if i < 1000 {
			// 	fmt.Print(byt, ",")
			// }
		}
		fmt.Println("breeze512", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		st = time.Now()
		for i := 0; i < n; i++ {
			bmap256.RandNorm()
			// if i < 1000 {
			// 	fmt.Print(rd, ",")
			// }
		}
		fmt.Println("breeze256 rdNorm", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		st = time.Now()
		for i := 0; i < n; i++ {
			byt = uint8(cmwc.rand_cmwc())
		}
		fmt.Println("cmwcRand", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		st = time.Now()
		for i := 0; i < n; i++ {
			bts[i] = slsa.nextByte()
			// if i < 1000 {
			// 	fmt.Print(slsa.nextByte(), ",")
			// }
		}
		fmt.Println("salsa", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op\n")

		// st = time.Now()
		// for i := 0; i < n; i++ {
		// 	lmap2.ByteMP(&byt)
		// }
		// fmt.Println("breeze64_32i mutex", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		// st = time.Now()
		// for i := 0; i < n; i++ {
		// 	bmap.ByteMP(&byt)
		// }
		// fmt.Println("breeze64_32 mutex", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		// st = time.Now()
		// for i := 0; i < n; i++ {
		// 	bmap72.ByteMP(&byt)
		// }
		// fmt.Println("breeze128_72 mutex", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		st = time.Now()
		for i := 0; i < n; i++ {
			bmap128.ByteMP(&byt)
		}
		fmt.Println("breeze128 mutex", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		st = time.Now()
		for i := 0; i < n; i++ {
			bmap256.ByteMP(&byt)
		}
		fmt.Println("breeze256 mutex", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		st = time.Now()
		for i := 0; i < n; i++ {
			bmap512.ByteMP(&byt)
		}
		fmt.Println("breeze512 mutex", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		st = time.Now()
		for i := 0; i < n; i++ {
			byt = uint8(cmwc.rand_cmwcMP())
		}
		fmt.Println("cmwcRand mutex", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

		st = time.Now()
		for i := 0; i < n; i++ {
			bts[i] = slsa.nextByteMP()
		}
		fmt.Println("salsa mutex", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op\n")
	}

	key := []byte("Andreas ISt Doof")
	if 1 == 1 {
		in, err := ioutil.ReadFile("./1627939364.bin")
		if err != nil {
			panic(1)
		}
		out := make([]byte, len(in))

		st = time.Now()
		// var lmap32 breeze.Breeze64_32
		// lmap32.XOR(&out, &in, &key)
		// fmt.Println("Breeze64_32 XOR:", time.Since(st).Seconds()/float64(1), "s/GB ", len(in))
		// fmt.Println(bytes.Equal(out, in))

		// st = time.Now()
		// lmap32.Reset()
		// lmap32.XOR(&out, &out, &key)
		// fmt.Println("Breeze64_32 XOR:", time.Since(st).Seconds()/float64(1), "s/GB")
		// fmt.Println(bytes.Equal(out, in))

		// st = time.Now()
		// var lmap72 breeze.Breeze128_72
		// lmap72.XOR(&out, &in, &key)
		// fmt.Println("Breeze128_72 XOR:", time.Since(st).Seconds()/float64(1), "s/GB ", len(in))
		// fmt.Println(bytes.Equal(out, in))

		// st = time.Now()
		// lmap72.Reset()
		// lmap72.XOR(&out, &out, &key)
		// fmt.Println("Breeze128_72 XOR:", time.Since(st).Seconds()/float64(1), "s/GB")
		// fmt.Println(bytes.Equal(out, in))

		// st = time.Now()
		var warmUP breeze.Breeze128
		warmUP.XOR(&out, &in, &key)
		//fmt.Println("Breeze128 XOR:", time.Since(st).Seconds()/float64(1), "s/GB ", len(in))
		//fmt.Println(bytes.Equal(out, in))

		// st = time.Now()
		warmUP.Reset()
		warmUP.XOR(&out, &out, &key)
		//fmt.Println("Breeze128 XOR:", time.Since(st).Seconds()/float64(1), "s/GB")
		//fmt.Println(bytes.Equal(out, in))

		// competition
		st = time.Now()
		var lmap128 breeze.Breeze128
		lmap128.XOR(&out, &in, &key)
		fmt.Println("Breeze128 XOR:", time.Since(st).Seconds()/float64(1), "s/GB ", len(in))
		fmt.Println(bytes.Equal(out, in))

		st = time.Now()
		lmap128.Reset()
		lmap128.XOR(&out, &out, &key)
		fmt.Println("Breeze128 XOR:", time.Since(st).Seconds()/float64(1), "s/GB")
		fmt.Println(bytes.Equal(out, in))

		st = time.Now()
		var lmap256 breeze.Breeze256
		lmap256.XOR(&out, &in, &key)
		fmt.Println("Breeze256 XOR:", time.Since(st).Seconds()/float64(1), "s/GB ", len(in))
		fmt.Println(bytes.Equal(out, in))

		st = time.Now()
		lmap256.Reset()
		lmap256.XOR(&out, &out, &key)
		fmt.Println("Breeze256 XOR:", time.Since(st).Seconds()/float64(1), "s/GB")
		fmt.Println(bytes.Equal(out, in))

		st = time.Now()
		var lmap512 breeze.Breeze512
		lmap512.XOR(&out, &in, &key)
		fmt.Println("Breeze512 XOR:", time.Since(st).Seconds()/float64(1), "s/GB ", len(in))
		fmt.Println(bytes.Equal(out, in))

		st = time.Now()
		lmap512.Reset()
		lmap512.XOR(&out, &out, &key)
		fmt.Println("Breeze512 XOR:", time.Since(st).Seconds()/float64(1), "s/GB")
		fmt.Println(bytes.Equal(out, in))

		// salsa
		st = time.Now()
		var salsa prnGenXORSalsa
		salsa.XOR(out, in, string(key))
		fmt.Println("salsa XOR:", time.Since(st).Seconds()/float64(1), "s/GB ", len(in))
		fmt.Println(bytes.Equal(out, in))

		st = time.Now()
		var salsa1 prnGenXORSalsa
		salsa1.XOR(out, out, string(key))
		fmt.Println("salsa XOR:", time.Since(st).Seconds()/float64(1), "s/GB ")

		fmt.Println(bytes.Equal(out, in))
	}

	n = 1000
	fmt.Println("\nHash Timings ( n =", n, ")")

	hkey := key
	// st = time.Now()
	// for i := 0; i < n; i++ {
	// 	var hmap32 breeze.Breeze64_32
	// 	hmap32.ShortHash(hkey, 128)
	// 	hkey = append(key, byte(i))
	// }
	// fmt.Println("breeze64_32 hash", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

	// hkey = key
	// st = time.Now()
	// for i := 0; i < n; i++ {
	// 	var hmap72 breeze.Breeze128_72
	// 	hmap72.ShortHash(hkey, 128)
	// 	hkey = append(key, byte(i))
	// }
	// fmt.Println("breeze128_72 hash", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

	// hkey = key
	st = time.Now()
	for i := 0; i < n; i++ {
		var hmap128 breeze.Breeze128
		hmap128.ShortHash(hkey, 128)
		hkey = append(key, byte(i))
	}
	fmt.Println("breeze128 hash", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

	hkey = key
	st = time.Now()
	for i := 0; i < n; i++ {
		var hmap256 breeze.Breeze256
		hmap256.ShortHash(hkey, 128)
		hkey = append(key, byte(i))
	}
	fmt.Println("breeze256 hash", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

	hkey = key
	st = time.Now()
	for i := 0; i < n; i++ {
		var hmap512 breeze.Breeze512
		hmap512.ShortHash(hkey, 128)
		hkey = append(key, byte(i))
	}
	fmt.Println("breeze512 hash", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

	hkey = key
	st = time.Now()
	for i := 0; i < n; i++ {
		_ = sipHash(hkey)
		hkey = append(key, byte(i))
	}
	fmt.Println("sipHash-2-4", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

	skey := string(key)
	st = time.Now()
	for i := 0; i < n; i++ {
		hash := md5.New()
		io.WriteString(hash, skey)
		_ = hash.Sum(nil)
		skey += "a"
	}
	fmt.Println("md5 hash", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

	skey = string(key)
	st = time.Now()
	for i := 0; i < n; i++ {
		hash := sha256.New()
		io.WriteString(hash, skey)
		_ = hash.Sum(nil)
		skey += "a"
	}
	fmt.Println("sha256 hash", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

	skey = string(key)
	st = time.Now()
	for i := 0; i < n; i++ {
		hash := sha512.New()
		io.WriteString(hash, skey)
		_ = hash.Sum(nil)
		skey += "a"
	}
	fmt.Println("sha512 hash", float64(time.Since(st).Nanoseconds())/float64(n), "ns/op")

}

// //
// // LogisticMap
// //
// type LogistMap struct {
// 	State    uint64
// 	State1   float64
// 	State2   float64
// 	seed     uint64
// 	bitshift uint8
// 	idx      uint8
// }

// func (l *LogistMap) Seed(s uint64) {
// 	switch s {
// 	case 0, 1, 2, 4:
// 		s++
// 	}
// 	l.seed = s
// 	l.State1 = 1.0 / float64(s)
// 	l.State2 = 1.0 / (1.0 + float64(s)/3.1457)
// 	for i := 0; i < 5; i++ {
// 		l.roundTrip()
// 	}
// }

// func (l *LogistMap) roundTrip() {
// 	// x = (x - x*x) * 4.0
// 	newState := 4.0 * l.State1 * (1.0 - l.State1)
// 	switch newState {
// 	case 0, 0.25:
// 		l.Seed(l.seed + 1)
// 	default:
// 		l.State1 = 1.0 - newState
// 	}
// 	newState = 4.0 * l.State2 * (1.0 - l.State2)
// 	switch newState {
// 	case 0, 0.25:
// 		l.Seed(l.seed + 1)
// 	default:
// 		l.State2 = 1.0 - newState
// 	}
// 	l.bitshift = (l.bitshift + 1) % 23
// 	// l.State ^= l.State
// 	l.State ^= (uint64)((*(*uint64)(unsafe.Pointer(&l.State1))) << 32)
// 	l.State ^= (uint64)((*(*uint64)(unsafe.Pointer(&l.State2)))<<11) + (*(*uint64)(unsafe.Pointer(&l.State1)))<<11>>(12+l.bitshift)
// 	l.State ^= ((uint64)((*(*uint64)(unsafe.Pointer(&l.State1)))<<11>>(12+l.bitshift)) ^ (uint64)((*(*uint64)(unsafe.Pointer(&l.State2)))<<11>>(11+l.bitshift)))
// }

// func (l *LogistMap) Byte(byt *uint8) {
// 	// l.roundTrip()
// 	*byt = (uint8)(*(*uint64)(unsafe.Pointer(&l.State)) >> l.idx)
// 	l.idx += 1
// 	if l.idx == 8 {
// 		l.roundTrip()
// 		l.idx = 0
// 	}
// }

// //
// // LogisticMap
// //
// type LogistMap2 struct {
// 	State    [4]uint64
// 	State1   float64
// 	State2   float64
// 	newState float64
// 	seed     uint64
// 	bitshift uint8
// 	idx      uint8
// 	strt     unsafe.Pointer
// 	r1, r2   float64
// }

// func (l *LogistMap2) Init(s interface{}) {
// 	l.r1 = 4.0
// 	l.r2 = 4.0 - 0.0000000001
// 	switch s := s.(type) {
// 	case int:
// 		if s < 0 {
// 			s = -s
// 		}
// 		l.seed = uint64(s)
// 	case int8:
// 		if s < 0 {
// 			s = -s
// 		}
// 		l.seed = uint64(s)
// 	case int16:
// 		if s < 0 {
// 			s = -s
// 		}
// 		l.seed = uint64(s)
// 	case int32:
// 		if s < 0 {
// 			s = -s
// 		}
// 		l.seed = uint64(s)
// 	case int64:
// 		if s < 0 {
// 			s = -s
// 		}
// 		l.seed = uint64(s)
// 	case uint8:
// 		l.seed = uint64(s)
// 	case uint16:
// 		l.seed = uint64(s)
// 	case uint32:
// 		l.seed = uint64(s)
// 	case uint64:
// 		l.seed = s
// 	case []byte:
// 		l.seed = uint64(s[0])
// 		hlp := uint64(1)
// 		for i := 1; i < len(s); i++ {
// 			hlp += uint64(s[i])
// 			l.seed += hlp
// 			hlp <<= 1
// 		}
// 	case string:
// 		l.seed = uint64(s[0])
// 		hlp := uint64(1)
// 		for i := 1; i < len(s); i++ {
// 			hlp += uint64(s[i])
// 			l.seed += hlp
// 			hlp <<= 1
// 		}
// 	case float64:
// 		l.seed = *(*uint64)(unsafe.Pointer(&s)) << 12 >> 12 // (uint64)(*(*uint64)(unsafe.Pointer(&s)))
// 	default:
// 		panic(1)
// 	}
// 	l.seedr()
// }

// func (l *LogistMap2) seedr() {
// 	var s1, s2 = uint64(1<<28 - 10), uint64(1<<28 - 10)
// 	// l.seed = l.seed << 13 >> 13
// 	for l.seed&1 == 0 {
// 		l.seed >>= 1
// 	}
// 	done := false
// 	for !done {
// 		for i := uint(63); i > 0; i-- {
// 			if l.seed>>i == 1 {
// 				s1 = (l.seed >> (i >> 1)) % s1
// 				s2 = (l.seed << (63 - (i >> 1)) >> (63 - (i >> 1))) % s2
// 				if s1 != s2 && s1 != (1<<28-10) && s2 != 0 {
// 					done = true
// 					break
// 				}
// 			}
// 		}
// 		l.seed = l.seed << 1 >> 1
// 		if l.seed == 0 {
// 			l.seed = s1
// 		}
// 	}

// 	// fmt.Printf("%v %v\n", s1, s2)

// 	switch s1 {
// 	case 0, 1, 2, 4:
// 		s1 += 5
// 	}

// 	switch s2 {
// 	case 0, 1, 2, 4:
// 		s2 += 7
// 	}

// 	l.State1 = 1.0 / float64(s1)
// 	l.State2 = 1.0 / float64(s2)

// 	// for i := 0; i < 5; i++ {
// 	// fmt.Println(i, l.State1, l.State2)
// 	l.roundTrip()
// 	// }
// 	l.strt = unsafe.Pointer(&l.State[0])

// }

// func (l *LogistMap2) roundTrip() {
// 	l.newState = l.r1 * l.State1 * (1.0 - l.State1)
// 	switch l.newState {
// 	case 0, 0.25:
// 		fmt.Print("1@", l.seed)
// 		l.seed = (uint64)((*(*uint64)(unsafe.Pointer(&l.State1)))<<11>>(12+l.bitshift%7)) | (uint64)((*(*uint64)(unsafe.Pointer(&l.State2)))<<11>>(12+l.bitshift%7))
// 		fmt.Print("->", l.seed)
// 		l.seedr()
// 	default:
// 		l.State1 = 1.0 - l.newState
// 	}
// 	l.newState = l.r2 * l.State2 * (1.0 - l.State2)
// 	switch l.newState {
// 	case 0:
// 		fmt.Print("2@", l.seed)
// 		l.seed = (uint64)((*(*uint64)(unsafe.Pointer(&l.State1)))<<11>>(12+l.bitshift%7)) | (uint64)((*(*uint64)(unsafe.Pointer(&l.State2)))<<11>>(12+l.bitshift%7))
// 		fmt.Print("->", l.seed)
// 		l.r2 -= 0.0000000001
// 		l.seedr()
// 	default:
// 		l.State2 = 1.0 - l.newState
// 	}

// 	l.bitshift = (l.bitshift + 1) % 22
// 	l.State[0] ^= (uint64)((*(*uint64)(unsafe.Pointer(&l.State1))) << 32)
// 	l.State[0] ^= (uint64)((*(*uint64)(unsafe.Pointer(&l.State2)))<<11) + (*(*uint64)(unsafe.Pointer(&l.State1)))<<11>>(12+l.bitshift)
// 	l.State[0] ^= ((uint64)((*(*uint64)(unsafe.Pointer(&l.State1)))<<11>>(12+l.bitshift)) ^ (uint64)((*(*uint64)(unsafe.Pointer(&l.State2)))<<11>>(13+l.bitshift)))

// 	l.bitshift++
// 	l.State[1] ^= (uint64)((*(*uint64)(unsafe.Pointer(&l.State2))) << 32)
// 	l.State[1] ^= (uint64)((*(*uint64)(unsafe.Pointer(&l.State1)))<<11) + (*(*uint64)(unsafe.Pointer(&l.State2)))<<11>>(12+l.bitshift)
// 	l.State[1] ^= ((uint64)((*(*uint64)(unsafe.Pointer(&l.State2)))<<11>>(12+l.bitshift)) ^ (uint64)((*(*uint64)(unsafe.Pointer(&l.State1)))<<11>>(13+l.bitshift)))

// 	l.State[2] ^= l.State[0]
// 	l.State[2] ^= (uint64)((*(*uint64)(unsafe.Pointer(&l.State2)))<<11) + (*(*uint64)(unsafe.Pointer(&l.State1)))<<11>>(12+l.bitshift)

// 	l.State[3] ^= l.State[1]
// 	l.State[3] ^= (uint64)((*(*uint64)(unsafe.Pointer(&l.State1)))<<11) + (*(*uint64)(unsafe.Pointer(&l.State2)))<<11>>(12+l.bitshift) ^ l.State[2]

// 	l.State[2] ^= l.State[3]

// }

// func (l *LogistMap2) Byte(byt *uint8) {
// 	*byt = (uint8)(*(*uint8)(unsafe.Pointer(uintptr(l.strt) + uintptr(l.idx))))
// 	l.idx++
// 	if l.idx == 32 {
// 		l.roundTrip()
// 		l.idx = 0
// 	}
// }

// func (l *LogistMap2) ByteMP(byt *uint8) {
// 	mutex.Lock()
// 	*byt = (uint8)(*(*uint8)(unsafe.Pointer(uintptr(l.strt) + uintptr(l.idx))))
// 	l.idx++
// 	if l.idx == 32 {
// 		l.roundTrip()
// 		l.idx = 0
// 	}
// 	mutex.Unlock()
// }
