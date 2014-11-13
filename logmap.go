package main

import (
	// "bytes"
	"bufio"
	"crypto/sha512"
	"fmt"
	"math/rand"
	"sync"
	// "unsafe"
	// "math/big"
	"io"
	"io/ioutil"
	"runtime"
	// "strings"
	"os"
	"time"

	"./breeze"
	"flag"
)

var (
	x                                     float64
	mutex                                 = &sync.Mutex{}
	swgrp                                 sync.WaitGroup
	NUM_scipps, NUM_repeats, NUM_stateDBL uint64
	module2test, outputFile               string
	testLenInBytes, rep                   int
	Hstring                               = "Do'nt be a fool! -(o o)-"
	Width                                 int
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

type repeatCheck struct {
	repeatLst map[float64]int
}

func (r *repeatCheck) init() {
	r.repeatLst = make(map[float64]int)
}

func (r *repeatCheck) add(v float64) {
	r.repeatLst[v]++
}

func (r *repeatCheck) output() {
	for _, v := range r.repeatLst {
		if v > Width {
			// fmt.Print(k, v, ", ")
			NUM_repeats++
		}
	}
	// fmt.Println()
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

func (s *prnGenSalsa) XOR(out, in []byte, keyPhrase string) {
	hash := sha512.New()
	io.WriteString(hash, keyPhrase)
	key := hash.Sum(nil)
	copy(s.keyBytes[:], key[:])
	// core(&(s.block), &(s.counter), &(s.keyBytes), &Sigma)
	xORKeyStream(out, in, &(s.counter), &(s.keyBytes))
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

////

func init() {
	flag.IntVar(&testLenInBytes, "l", 1000000, "No of bytes to check (default 1000000)")
	flag.IntVar(&rep, "r", 10, "No of repetions (default 10)")
	flag.StringVar(&module2test, "m", "-", "one of 128 | 256 | 512 | CS128")
	flag.StringVar(&outputFile, "o", "", "any input produces output file with following name scheme: 'module_randomNumber.bin'; omit for no output")
}

func main() {
	flag.Parse()

	if module2test == "-" {
		fmt.Println("You need to declare which breeze module should be tested. See --h for options.")
		return
	}

	rand.Seed(int64(time.Now().Nanosecond()))

	var st = time.Now() // starttime

	// byte baskets
	var bskCompLst float64

	// output register
	var (
		outBin = make([]byte, testLenInBytes*rep)
		oi     = 0
	)

	// output state checking
	// var ergs = [rep][17]uint64{}

	for ii := 0; ii < rep; ii++ {

		// var sPrng prnGenSalsa
		// sPrng.Init(hstring)
		// Hstring += " say it again "

		var listerLen = 6
		Width = 128

		switch module2test {
		case "256":
			listerLen = 12
			Width = 256
		case "512":
			listerLen = 24
			Width = 512
		}

		var lst = make([]repeatCheck, listerLen)
		for i, _ := range lst {
			lst[i].init()
		}

		// rlst0 := make(map[uint64]int)
		// rlst1 := make(map[uint64]int)
		// rlst2 := make(map[uint64]int)
		// rlst3 := make(map[uint64]int)
		// rlst4 := make(map[uint64]int)
		// rlst5 := make(map[uint64]int)
		// rlst6 := make(map[uint64]int)
		// rlst7 := make(map[uint64]int)
		// rlst8 := make(map[uint64]int)
		// rlst9 := make(map[uint64]int)
		// rlst10 := make(map[uint64]int)
		// rlst11 := make(map[uint64]int)
		// rlst12 := make(map[uint64]int)
		// rlst13 := make(map[uint64]int)
		// rlst14 := make(map[uint64]int)
		// rlst15 := make(map[uint64]int)

		var bsk = make([]uint64, 256)
		var byt uint8

		switch module2test {
		case "-":
			fmt.Println("You need to declare which breeze module should be tested. See --h for options.")
			return
		case "128":
			var lmap breeze.Breeze128
			lmap.Init(rand.Int63())
			// lmap.Init(Hstring)

			for i := 0; i < testLenInBytes; i++ {
				//for iii := 0; iii < Width; iii++ {
				lmap.Byte(&byt)
				//}
				lst[0].add(lmap.State1)
				lst[1].add(lmap.State2)
				lst[2].add(lmap.State3)
				lst[3].add(lmap.State4)
				lst[4].add(lmap.State5)
				lst[5].add(lmap.State6)
				// if i%Width == 0 {
				bsk[byt]++
				// }
				outBin[oi] = byt //byte(sPrng.nextByte())
				oi++
			}
		case "CS128":
			var lmap breeze.BreezeCS128
			lmap.Init()
			// lmap.Init(Hstring)

			for i := 0; i < testLenInBytes; i++ {
				//for iii := 0; iii < Width; iii++ {
				lmap.Byte(&byt)
				//}
				lst[0].add(lmap.State1)
				lst[1].add(lmap.State2)
				lst[2].add(lmap.State3)
				lst[3].add(lmap.State4)
				lst[4].add(lmap.State5)
				lst[5].add(lmap.State6)
				// if i%Width == 0 {
				bsk[byt]++
				// }
				outBin[oi] = byt //byte(sPrng.nextByte())
				oi++
			}
		case "256":
			var lmap breeze.Breeze256
			lmap.Init([]uint64{uint64(rand.Int63()), uint64(rand.Int63())})
			// lmap.Init(Hstring)

			for i := 0; i < testLenInBytes; i++ {
				//for iii := 0; iii < Width; iii++ {
				lmap.Byte(&byt)
				//}
				lst[0].add(lmap.State1)
				lst[1].add(lmap.State2)
				lst[2].add(lmap.State3)
				lst[3].add(lmap.State4)
				lst[4].add(lmap.State5)
				lst[5].add(lmap.State6)

				lst[6].add(lmap.State7)
				lst[7].add(lmap.State8)
				lst[8].add(lmap.State9)
				lst[9].add(lmap.State10)
				lst[10].add(lmap.State11)
				lst[11].add(lmap.State12)
				// if i%Width == 0 {
				bsk[byt]++
				// }
				outBin[oi] = byt //byte(sPrng.nextByte())
				oi++
			}
		case "512":
			var lmap breeze.Breeze512
			lmap.Init([]uint64{uint64(rand.Int63()), uint64(rand.Int63())})
			// lmap.Init(Hstring)

			for i := 0; i < testLenInBytes; i++ {
				//for iii := 0; iii < Width; iii++ {
				lmap.Byte(&byt)
				//}
				lst[0].add(lmap.State1)
				lst[1].add(lmap.State2)
				lst[2].add(lmap.State3)
				lst[3].add(lmap.State4)
				lst[4].add(lmap.State5)
				lst[5].add(lmap.State6)

				lst[6].add(lmap.State7)
				lst[7].add(lmap.State8)
				lst[8].add(lmap.State9)
				lst[9].add(lmap.State10)
				lst[10].add(lmap.State11)
				lst[11].add(lmap.State12)

				lst[12].add(lmap.State1)
				lst[13].add(lmap.State2)
				lst[14].add(lmap.State3)
				lst[15].add(lmap.State4)
				lst[16].add(lmap.State5)
				lst[17].add(lmap.State6)
				lst[18].add(lmap.State7)
				lst[19].add(lmap.State8)
				lst[20].add(lmap.State9)
				lst[21].add(lmap.State10)
				lst[22].add(lmap.State11)
				lst[23].add(lmap.State12)
				// if i%Width == 0 {
				bsk[byt]++
				// }
				outBin[oi] = byt //byte(sPrng.nextByte())
				oi++
			}
		}

		// lmap.Init([]uint64{uint64(rand.Int63()), uint64(rand.Int63())})
		// lmap.Init(Hstring)

		// var st = time.Now()
		// for i := 0; i < testLenInBytes; i++ {
		// 	//for iii := 0; iii < Width; iii++ {
		// 	lmap.Byte(&byt)
		// 	//}
		// 	lst[1].add(lmap.State1)
		// 	lst[2].add(lmap.State2)
		// 	lst[3].add(lmap.State3)
		// 	lst[4].add(lmap.State4)
		// 	lst[5].add(lmap.State5)
		// 	lst[6].add(lmap.State6)
		// 	if module2test == "256" || module2test == "512" {
		// 		lst[7].add(lmap.State7)
		// 		lst[8].add(lmap.State8)
		// 		lst[9].add(lmap.State9)
		// 		lst[10].add(lmap.State10)
		// 		lst[11].add(lmap.State11)
		// 		lst[12].add(lmap.State12)
		// 	}
		// 	if module2test == "512" {
		// 		lst[13].add(lmap.State1)
		// 		lst[14].add(lmap.State2)
		// 		lst[15].add(lmap.State3)
		// 		lst[16].add(lmap.State4)
		// 		lst[17].add(lmap.State5)
		// 		lst[18].add(lmap.State6)

		// 		lst[19].add(lmap.State7)
		// 		lst[20].add(lmap.State8)
		// 		lst[21].add(lmap.State9)
		// 		lst[22].add(lmap.State10)
		// 		lst[23].add(lmap.State11)
		// 		lst[24].add(lmap.State12)
		// 	}
		// rlst0[lmap.State[0]]++
		// rlst1[lmap.State[1]]++
		// rlst2[lmap.State[2]]++
		// rlst3[lmap.State[3]]++
		// rlst4[lmap.State[4]]++
		// rlst5[lmap.State[5]]++
		// rlst6[lmap.State[6]]++
		// rlst7[lmap.State[7]]++
		// rlst8[lmap.State[8]]++
		// rlst9[lmap.State[9]]++
		// rlst10[lmap.State[10]]++
		// rlst11[lmap.State[11]]++
		// rlst12[lmap.State[12]]++
		// rlst13[lmap.State[13]]++
		// rlst14[lmap.State[14]]++
		// rlst15[lmap.State[15]]++

		// 	// if i%Width == 0 {
		// 	bsk[byt]++
		// 	// }

		// 	outBin[oi] = byt //byte(sPrng.nextByte())
		// 	oi++

		// }

		// for k, _ := range lst1.repeatLst {
		// 	if _, okay := lst2.repeatLst[k]; okay {
		// 		//fmt.Println("dbl:", k, "  ")
		// 		NUM_stateDBL++
		// 	}
		// 	if _, okay := lst3.repeatLst[k]; okay {
		// 		//fmt.Println("dbl:", k, "  ")
		// 		NUM_stateDBL++
		// 	}
		// 	if _, okay := lst4.repeatLst[k]; okay {
		// 		//fmt.Println("dbl:", k, "  ")
		// 		NUM_stateDBL++
		// 	}
		// }
		// for k, _ := range lst2.repeatLst {
		// 	if _, okay := lst3.repeatLst[k]; okay {
		// 		//fmt.Println("dbl:", k, "  ")
		// 		NUM_stateDBL++
		// 	}
		// 	if _, okay := lst4.repeatLst[k]; okay {
		// 		//fmt.Println("dbl:", k, "  ")
		// 		NUM_stateDBL++
		// 	}
		// }
		// for k, _ := range lst3.repeatLst {
		// 	if _, okay := lst4.repeatLst[k]; okay {
		// 		//fmt.Println("dbl:", k, "  ")
		// 		NUM_stateDBL++
		// 	}
		// }

		fmt.Print("inner state repeats: ")
		for i := 0; i < listerLen; i++ {
			fmt.Printf("(%v) ", i+1)
			lst[i].output()
		}

		// fmt.Print("[0]")
		// for k, v := range rlst0 {
		// 	if v > width { //32 {
		// 		fmt.Print(k, v, ", ")
		// 		NUM_repeats++
		// 	}
		// }
		// // fmt.Println()
		// fmt.Print("[1]")
		// for k, v := range rlst1 {
		// 	if v > width { //32 {
		// 		fmt.Print(k, v, ", ")
		// 		NUM_repeats++
		// 	}
		// }
		// // fmt.Println()
		// fmt.Print("[2]")
		// for k, v := range rlst2 {
		// 	if v > width { //32 {
		// 		fmt.Print(k, v, ", ")
		// 		NUM_repeats++
		// 	}
		// }
		// // fmt.Println()
		// fmt.Print("[3]")
		// for k, v := range rlst3 {
		// 	if v > width { //32 {
		// 		fmt.Print(k, v, ", ")
		// 		NUM_repeats++
		// 	}
		// }
		// fmt.Print("[4]")
		// for k, v := range rlst4 {
		// 	if v > width { //32 {
		// 		fmt.Print(k, v, ", ")
		// 		NUM_repeats++
		// 	}
		// }
		// // fmt.Println()
		// fmt.Print("[5]")
		// for k, v := range rlst5 {
		// 	if v > width { //32 {
		// 		fmt.Print(k, v, ", ")
		// 		NUM_repeats++
		// 	}
		// }
		// // fmt.Println()
		// fmt.Print("[6]")
		// for k, v := range rlst6 {
		// 	if v > width { //32 {
		// 		fmt.Print(k, v, ", ")
		// 		NUM_repeats++
		// 	}
		// }
		// // fmt.Println()
		// fmt.Print("[7]")
		// for k, v := range rlst7 {
		// 	if v > width { //32 {
		// 		fmt.Print(k, v, ", ")
		// 		NUM_repeats++
		// 	}
		// }
		// // fmt.Println()
		// fmt.Print("[8]")
		// for k, v := range rlst8 {
		// 	if v > width { //32 {
		// 		fmt.Print(k, v, ", ")
		// 		NUM_repeats++
		// 	}
		// }
		// // fmt.Println()
		// fmt.Print("[9]")
		// for k, v := range rlst1 {
		// 	if v > width { //32 {
		// 		fmt.Print(k, v, ", ")
		// 		NUM_repeats++
		// 	}
		// }
		// // fmt.Println()
		// fmt.Print("[10]")
		// for k, v := range rlst2 {
		// 	if v > width { //32 {
		// 		fmt.Print(k, v, ", ")
		// 		NUM_repeats++
		// 	}
		// }
		// // fmt.Println()
		// fmt.Print("[11]")
		// for k, v := range rlst3 {
		// 	if v > width { //32 {
		// 		fmt.Print(k, v, ", ")
		// 		NUM_repeats++
		// 	}
		// }
		// fmt.Print("[12]")
		// for k, v := range rlst4 {
		// 	if v > width { //32 {
		// 		fmt.Print(k, v, ", ")
		// 		NUM_repeats++
		// 	}
		// }
		// // fmt.Println()
		// fmt.Print("[13]")
		// for k, v := range rlst5 {
		// 	if v > width { //32 {
		// 		fmt.Print(k, v, ", ")
		// 		NUM_repeats++
		// 	}
		// }
		// // fmt.Println()
		// fmt.Print("[14]")
		// for k, v := range rlst6 {
		// 	if v > width { //32 {
		// 		fmt.Print(k, v, ", ")
		// 		NUM_repeats++
		// 	}
		// }
		// // fmt.Println()
		// fmt.Print("[15]")
		// for k, v := range rlst7 {
		// 	if v > width { //32 {
		// 		fmt.Print(k, v, ", ")
		// 		NUM_repeats++
		// 	}
		// }

		fmt.Println(" innerState dbl:", NUM_stateDBL) //, "  outStates dbl:", NUM_repeats)
		// // 	for ii := i + 1; ii < len(lst)-i; ii++ {
		// // 		if v == lst[ii] {
		// // 			fmt.Println(i, v, ii, lmap)
		// // 			NUM_repeats++
		// // 		}
		// // 	}
		// // }

		fmt.Println(bsk)
		var bskcomp, bskcomp1, bskcomp2 float64
		for i := 0; i < 128; i++ {
			bskcomp += float64(bsk[i]) / float64(bsk[255-i])
			bskcomp1 += float64(bsk[i])
			bskcomp2 += float64(bsk[255-i])
		}
		bskCompLst += (bskcomp / 128.0)
		bskCompLst += (bskcomp1 / bskcomp2)
		fmt.Println((bskcomp1 / bskcomp2), (bskcomp / 128.0))

		// l := 6
		// for s := l; s < n-l; s++ {
		// 	// fmt.Print("start:", s, ": ")
		// 	for i := l; i < n; i++ {
		// 		if bytes.Equal(lst[s:s+l], lst[i-l:i]) {
		// 			if i != s+l {
		// 				fmt.Print(i, "-->", lst[s:s+l], "@[", s, "]\n")
		// 			}
		// 		}
		// 	}
		// 	// fmt.Println()
		// }

		// for ii := 0; ii < 16; ii++ {
		// 	ergs[c-start][ii] = lmap.State[ii]
		// }
		// ergs[c-start][16] = uint64(rr)

	}
	// for i := 0; i < rep; i++ {
	// 	for ii := i + 1; ii < rep; ii++ {
	// 		if ergs[i][0] == ergs[ii][0] && ergs[i][1] == ergs[ii][1] && ergs[i][2] == ergs[ii][2] && ergs[i][3] == ergs[ii][3] {
	// 			if ergs[i][4] == ergs[ii][4] && ergs[i][5] == ergs[ii][5] && ergs[i][6] == ergs[ii][6] && ergs[i][7] == ergs[ii][7] && ergs[i][8] == ergs[ii][8] {
	// 				fmt.Println(i, ii, ergs[i], "= rr:", ergs[ii][9])
	// 			}
	// 		}
	// 	}
	// }

	fmt.Println("lmap:", time.Since(st).Nanoseconds()/int64(testLenInBytes*rep)) // 500))
	fmt.Println("tests:", rep, "length:", testLenInBytes)
	fmt.Println("repeats:", NUM_repeats, "\nstateDbl:", NUM_stateDBL, "  baskets:", bskCompLst/float64(2*rep))

	if outputFile != "" {
		nm := rand.Int31()
		if err := ioutil.WriteFile(fmt.Sprintf("%v_%v.bin", module2test, nm), outBin, 0666); err != nil {
			fmt.Println("NO SUCCESS")
			panic(0)
		}
		fmt.Printf("output to %v_%v.bin (%v sequences with length %v Bytes)\n", module2test, nm, rep, testLenInBytes)
	}

	// in, err := ioutil.ReadFile("./418047969.bin")
	// if err != nil {
	// 	panic(1)
	// }
	// out := make([]byte, len(in))
	// st = time.Now()
	// var lmap2 breeze.Breeze512
	// lmap2.XOR(&out, &in, &Hstring)
	// fmt.Println("lmap512 XOR:", time.Since(st).Seconds(), "s/GB ", len(in))
	// fmt.Println(bytes.Equal(out, in))

	// st = time.Now()
	// lmap2.Reset()
	// lmap2.XOR(&out, &out, &Hstring)
	// fmt.Println("lmap512 XOR:", time.Since(st).Seconds(), "s/GB")
	// fmt.Println(bytes.Equal(out, in))

	// // salsa
	// st = time.Now()
	// var salsa prnGenSalsa
	// salsa.XOR(out, in, string(Hstring))
	// fmt.Println("salsa XOR:", time.Since(st).Seconds(), "s/GB ", len(in))
	// fmt.Println(bytes.Equal(out, in))

	// st = time.Now()
	// var salsa1 prnGenSalsa
	// salsa1.XOR(out, out, string(Hstring))
	// fmt.Println("salsa XOR:", time.Since(st).Seconds(), "s/GB ")

	// fmt.Println(bytes.Equal(out, in))

	var ww = make(map[string][]string)
	if 1 == 0 {
		f, err := os.Open("~/virtualSSH/goethe/worte_de_all.txt")
		if err != nil {
			panic(1)
		}
		var lmap3 breeze.Breeze256
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			lmap3.Reset()
			t := scanner.Text()
			h, _ := lmap3.ShortHash(t, 64/8)
			// fmt.Printf("%x : %v\n", h, t)
			ww[string(h)] = append(ww[string(h)], t)
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading file:", err)
		}

	}
	i := 0
	for k, v := range ww {
		if len(v) > 1 {
			i++
			fmt.Printf("%x: %v (%v) \n", k, v, i)
		}
	}

	// st = time.Now()
	// for i := 0; i < testLenInBytes; i++ {
	// 	sPrng.nextByte()
	// }
	// fmt.Println("salsa:", time.Since(st).Nanoseconds()/int64(n))
}
