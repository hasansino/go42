package tools

/*
	| Buffer Size (`2^n`) | Elements | Memory per Slot | Total Memory |
	|---------------------|----------|-----------------|--------------|
	| `2^10`              | 1,024    | 32 bytes        | 32 KB        |
	| `2^11`              | 2,048    | 32 bytes        | 64 KB        |
	| `2^12`              | 4,096    | 32 bytes        | 128 KB       |
	| `2^13`              | 8,192    | 32 bytes        | 256 KB       |
	| `2^14`              | 16,384   | 32 bytes        | 512 KB       |
	| `2^15`              | 32,768   | 32 bytes        | 1 MB         |
	| `2^16`              | 65,536   | 32 bytes        | 2 MB         |
*/

const (
	BufferSize1024  = 2 << 10
	BufferSize2048  = 2 << 11
	BufferSize4096  = 2 << 12
	BufferSize8192  = 2 << 13
	BufferSize16384 = 2 << 14
	BufferSize32768 = 2 << 15
	BufferSIze65536 = 2 << 16
)
