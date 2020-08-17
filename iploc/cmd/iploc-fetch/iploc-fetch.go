package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	flag "github.com/spf13/pflag"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

const (
	urlCopywrite = "http://update.cz88.net/ip/copywrite.rar"
	urlQqwryDat  = "http://update.cz88.net/ip/qqwry.rar"
)

var (
	outputFile string
	quiet, help bool

	fileSize  uint64
	fileSizeW int
)

type FetchWriter struct {
	n uint64
}

func (w *FetchWriter) Write(p []byte) (int, error) {
	n := len(p)
	w.n += uint64(n)
	w.progress()
	return n, nil
}

func (w *FetchWriter) progress() {
	printf("\rfetch: %6.2f%% %*d/%d bytes", float64(w.n)/float64(fileSize)*100, fileSizeW, w.n, fileSize)
}

func fetch(key uint32) error {
	resp, err := http.Get(urlQqwryDat)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	size, err := strconv.ParseUint(resp.Header.Get("Content-Length"), 10, 32)
	if err != nil {
		return err
	}
	if size != fileSize {
		return fmt.Errorf("the file size of the agreement is different")
	}

	var b []byte
	buffer := bytes.NewBuffer(b[:])

	_, err = io.Copy(buffer, io.TeeReader(resp.Body, &FetchWriter{}))
	printf("\n")
	if err != nil {
		return err
	}
	b = buffer.Bytes()

	// 解码前512字节
	for i := 0; i < 0x200; i++ {
		key *= 0x805
		key++
		key &= 0xFF
		b[i] = b[i] ^ byte(key)
	}

	r, err := zlib.NewReader(bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	defer r.Close()

	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	return err
}

func init() {
	flag.BoolVarP(&quiet, "quiet", "q", false, "only output error")
	flag.BoolVarP(&help, "help", "h", false, "this help")
	flag.CommandLine.SortFlags = false
	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "iploc-fetch: fetch qqwry.dat.\nUsage: iploc-fetch [output filename] [arguments]\nOptions:\n")
		flag.PrintDefaults()
	}

	outputFile = flag.Arg(0)

	if outputFile == "" || help {
		flag.Usage()
		if help {
			os.Exit(0)
		}
		os.Exit(1)
	}
}

func main() {
	resp, err := http.Get(urlCopywrite)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	fatal(err)

	s := readVersion(b)
	if s == nil {
		fatal(fmt.Errorf("invalid file description"))
	}
	fileSize = uint64(binary.LittleEndian.Uint32(b[12:]))
	fileSizeW = len(fmt.Sprint(fileSize))
	key := binary.LittleEndian.Uint32(b[20:])

	printf("version: %s\n", toUTF8(s))
	printf("fetch: ...")

	fatal(fetch(key))
}

func readVersion(p []byte) []byte {
	var start = 24
	var end int

	if start >= len(p) {
		return nil
	}

	// 0x20 ASCII space
	for p[start] != 0x20 {
		start++
	}
	start += 1
	end = start

	if end >= len(p) {
		return nil
	}

	for p[end] != 0x00 {
		end++
	}
	return p[start:end]
}

func toUTF8(s []byte) (b []byte) {
	r := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	b, _ = ioutil.ReadAll(r)
	return
}

func fatal(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error() + "\n")
		os.Exit(1)
	}
}

func printf(format string, args ...interface{}) {
	if quiet {
		return
	}
	fmt.Printf(format, args...)
}
