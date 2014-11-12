####breezeTest


Tests used when developing breeze (P)RNG ( see: https://github.com/AndreasBriese/breeze )


**logmap.go**

testing innerStates of breeze variants. An appropriate breeze draw-up with exposed states can be found in the breeze folder.

calls breeze.Breeze128/CS128/256/512, checks for innerState duplicates and pipes bytes to an output File

run '$ go run logmap.go salsa20.go -h ' to see further options



**randomImager.go randomPattern.html randPad.png**

runs over a file with outputBytes (*.bin) from logmap.go and converts 4 bytes into RGBA png (4th Byte -> Alpha).
will NOT stop automatically - make sure you stop it after a while.

	$ go run logmap.go salsa20.go -m="128" -o="yes" 
	## will produce a file 128_xxxxxx.bin with 10 times 1000000 bytes (seeded from rand.Int63())
    $ go run randomImager.go -f "128_xxxxxx.bin"

Now point your browser at randomPattern.html to inspect.   



**prngComp.go**

Starts a timing comparison

    $ go run prngComp.go salsa20.go sipHash.go 


