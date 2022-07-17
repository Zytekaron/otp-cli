# otp-cli

A simple [One Time Pad](https://en.wikipedia.org/wiki/One-time_pad)
command-line utility for generating and applying single-use key file encryption.

## Installation & Usage

### Building

- Install [Go](https://go.dev/dl), version 1.19 or later

```sh
go build -o otpc main.go # Windows: use 'otpc.exe'
mv otpc /usr/local/bin # Windows: move it to a folder in your PATH
```

### Usage

```
Generic
  -o, --output string     the file to write the output to
  -v, --verbose           enable verbose output
Generating Keys
  -g, --generate string   the size of the key file to generate
  -c, --count int         the number of key files to generate (default 1)
  -O, --offset int        the index offset (start) for the file index
Encryption & Decryption
  -k, --key string        the file containing the key
  -i, --input string      the file containing the input, else stdin
```

#### Generating Keys

Generate random key files using a cryptographically-secure random number source.
Roughly equivalent to running one of the following commands for a 1 GB key file:
- `head -c 1G </dev/urandom >key.txt`
- `dd if=/dev/urandom of=key.txt bs=1M count=1000` (1 MB x 1000)

```sh
# generating a single key
otpc -g 1GiB -o key.txt
# generating multiple keys (use quotes when using '{i}')
otpc -g 1024 -c 10 -o "keys/key-{i}.txt" # key-0.txt through key-9.txt
# generating ones you missed
otpc -g 1kb -c 5 -O 10 -o "keys/key-{i}.txt" # key-10.txt through key-14.txt
```

#### Encrypting and Decrypting Files

When encrypting and decrypting, the only requirement
is that the key is equal to or larger than the message.

```sh
# encrypting using files
otpc -k "key.txt" -i "message.txt" -o "encrypted.txt"
# decrypting using files
otpc -k "key.txt" -i "encrypted.txt" -o "message.txt"
# encrypting using stdin and stdout
echo "Hello, World" | otpc -k "key.txt" > encrypted.txt
cat encrypted.txt | otpc -k "key.txt" | tee message.txt
```

## About One Time Pads

Pros:
- Mathematically unbreakable encryption.
- Extremely simple to implement and verify the algorithm.
- Encryption is very fast, since it only requires XOR'ing two files together once.

Cons:
- Key files must be transferred offline, otherwise you are now relying on less secure encryption algorithms.
- Each message byte requires a unique key byte (not good for transferring lots of data).
- Keys cannot be safely re-used.

Usage Tips:
- [Never re-use keys for multiple messages.](https://crypto.stackexchange.com/a/108)
- Always delete keys once you have used them to encrypt or decrypt a message.
  Use a program like [srm](https://man7.org/linux/man-pages/man1/shred.1.html) or [shred](https://linux.die.net/man/1/shred)
  (caution: [these may not work on journaled filesystems](https://stackoverflow.com/a/913360))
- Generate keys of different sizes, and use the smallest one possible for a given message.*
- Compress large files using a program like [gzip](https://linux.die.net/man/1/gzip), so you can use smaller keys.
- Store your keys in an encrypted volume, such as a [VeraCrypt volume](https://en.wikipedia.org/wiki/VeraCrypt),
  until you need to use them. This also lets you use a non-journaled filesystem, so srm and shred will work.

[*] This program does not support encrypting multiple messages over time using a single key file,
unless you modify the file yourself to remove only the key bytes used for a previous operation.

That said, here's a way you can do that. **Only run these commands if you know what you are doing.
See the [dd man page](https://linux.die.net/man/1/dd) for more information.**

This command will remove the first 64 bytes of `key.txt`, placing the result in `stripped_key.txt`.

`dd if=key.txt of=stripped_key.txt ibs=64 skip=1`

If you need to strip a larger value (perhaps more than 50 MB), use a smaller value for `ibs`, and
a value for `skip` such that `ibs * skip = number of bytes to remove`. The following command will
do the same thing, removing the first 500 million bytes (500 MB) of the file, in 1 MB chunks.

`dd if=key.txt of=stripped_key.txt ibs=1000 skip=500`

## License

**otp-cli** is licensed under the [MIT License](./LICENSE)
