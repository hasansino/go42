package tools

/*
	| Buffer Size (`2^n`) | Elements | Total Memory |
	|---------------------|----------|--------------|
	| `2<<10`             | 1,024    | 64 KB        |
	| `2<<11`             | 2,048    | 128 KB       |
	| `2<<12`             | 4,096    | 256 KB       |
	| `2<<13`             | 8,192    | 512 KB       |
	| `2<<14`             | 16,384   | 1 MB         |
	| `2<<15`             | 32,768   | 2 MB         |
	| `2<<16`             | 65,536   | 4 MB         |
*/

const (
	BufferSize1024  = 2 << 10
	BufferSize2048  = 2 << 11
	BufferSize4096  = 2 << 12
	BufferSize8192  = 2 << 13
	BufferSize16384 = 2 << 14
	BufferSize32768 = 2 << 15
	BufferSize65536 = 2 << 16
)
