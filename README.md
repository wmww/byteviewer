# Byteviewer
CLI tool for viewing data in various decodings side-by-side.

## Installation
To build the program and install it on your local system:
```
go build -o byteviewer ./main
```

(Optional) install on your system:
```
sudo mv byteviewer /usr/local/bin
```

## Usage examples
Input can be piped from stdin:

```bash
$ cat myfile | byteviewer -i32 -hex -u8
position  u8                               i32                      hex                      
---------------------------------------------------------------------------------------------
       0  184, 21, 76,217, 80,210,138,124   -649325128, 2089472592  b8,15,4c,d9,50,d2,8a,7c  
       8  103,216,  7,229,185,202, 96, 76   -452470681, 1281411769  67,d8,07,e5,b9,ca,60,4c  
      16  211,192,125, 24, 36,111, 54, 64    410894547, 1077309220  d3,c0,7d,18,24,6f,36,40  
      24   97,155,172,230,218,224,199,167   -424895647,-1480072998  61,9b,ac,e6,da,e0,c7,a7  
      32  134,142,123,186,226,250, 79,210  -1166307706, -766510366  86,8e,7b,ba,e2,fa,4f,d2  
      40   43,159, 54,227,123, 17,141,153   -482959573,-1718808197  2b,9f,36,e3,7b,11,8d,99  
      48  126,179,218,121, 33,143, 87,241   2044375934, -245919967  7e,b3,da,79,21,8f,57,f1  
      56  212,102,233,255,171,138,154,176     -1481004,-1332049237  d4,66,e9,ff,ab,8a,9a,b0  
...
```

Or read from a file:
```bash
$ byteviewer -hex -utf8 -f32 -f myfile
```

You can view a subset with the start (-s) and length (-l) flags:
```bash
$ cat myfile | byteviewer -s 200 -l 100
```

You can view a subsection of the file with the start (-s) and length (-l) flags:
```bash
$ cat myfile | byteviewer -s 40 -l 16
```

And change the width in bytes with the -w flag:
```bash
$ cat myfile | byteviewer -w 4 -f32
position  f32           
------------------------
       0   -3.5903e+15  
       4   5.76642e+36  
       8  -4.00945e+22  
      12   5.89278e+07  
      16   3.27968e-24  
      20       2.85053  
...
```

In case the data isn't aligned with the file start, you can view multi-byte data types at all possible offsets with the -o flag.
```bash
cat myfile | ./byteviewer -w 4 -f32 -o
position  f32                                                  
---------------------------------------------------------------
       0   -3.5903e+15, 2.91651e+10, -2.2425e+11,-2.02527e-32  
       4   5.76642e+36,  1.1926e+24,-1.01809e+15, 3.25609e-34  
       8  -4.00945e+22,-0.000436841, -6.0914e+06, 1.16864e+20  
...
```

By default the output is printed in rainbow colors that change every two bytes, but you can disable that with -C or change the width of each color with -cw.

## Supported formats
- i8
- u8
- i16
- u16
- i32
- u32
- f32
- i64
- u64
- f64
- hex
- ascii
- utf8
- utf8h (hex formatted code points)
- utf8i (int formatted code points)
