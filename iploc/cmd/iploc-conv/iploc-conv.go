package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	flag "github.com/spf13/pflag"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"

	"github.com/kayon/iploc"
)

// 将原版 qqwry.dat 由 GBK 转换为 UTF-8
// 修正原 qqwry.dat 中几处错误的重定向，并将 "CZ88.NET" 替换为 "N/A"

var (
	qqwrySrc, qqwryDst   string
	quiet, noCheck, help bool
	gbkDecoder           transform.Transformer
)

// 修正 qqwry.dat 2018-05-10
var fix = map[uint32][2]string{
	3413565439: {"广东省广州市", "有线宽带"}, // 203.118.223.255
	3524192255: {"广东省广州市", "联通"},   // 210.14.231.255
	3526640127: {"福建省厦门市", "联通"},   // 210.52.65.255
	3526641407: {"广东省广州市", "联通"},   // 210.52.70.255
	3526646783: {"广东省佛山市", "联通"},   // 210.52.91.255
	3549601791: {"北京市", "广电网"},     // 211.146.159.255
	3549605631: {"北京市", "广电网"},     // 211.146.174.255
}

func init() {
	flag.StringVarP(&qqwrySrc, "src", "s", "", "source DAT file (GBK)")
	flag.StringVarP(&qqwryDst, "dst", "d", "", "destination DAT file (UTF-8)")
	flag.BoolVarP(&quiet, "quiet", "q", false, "only output error")
	flag.BoolVarP(&noCheck, "nc", "n", false, "do not check the converted DAT")
	flag.BoolVarP(&help, "help", "h", false, "this help")
	flag.CommandLine.SortFlags = false
	flag.Parse()

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "iploc-conv: converts DAT file from GBK into UTF-8.\nUsage: iploc-conv -s src.gbk.dat -d dst.utf8.dat [arguments]\nOptions:\n")
		flag.PrintDefaults()
	}

	if qqwrySrc == "" || qqwryDst == "" || help {
		flag.Usage()
		if help {
			os.Exit(0)
		}
		os.Exit(1)
	} else if qqwrySrc == qqwryDst {
		fmt.Fprintln(os.Stderr, "source and destination files is same")
		os.Exit(1)
	}

	gbkDecoder = simplifiedchinese.GBK.NewDecoder()
}

func main() {
	st := time.Now()
	parser, err := iploc.NewParser(qqwrySrc, true)
	if err != nil {
		fmt.Println(err)
		return
	}

	f, err := os.Create(qqwryDst)
	if err != nil {
		fmt.Println(err)
		return
	}

	var (
		raw              iploc.LocationRaw
		locOffset        = make(map[uint32]uint32)
		locPos           = make(map[string]uint32)
		indexes, s []byte
		head             [8]byte
		offset, o        uint32
		has              bool
		count            = parser.Count()
		l                = len(fmt.Sprint(count))
		w, n                int
		wt = bufio.NewWriter(f)
	)

	// write head 8 bytes
	wt.Write(head[:])

	// head 8 bytes
	offset = 8
	parser.IndexRange(func(i int, start, end, pos uint32) bool {
		if !quiet {
			fmt.Printf("\rBuild %6.2f%% %*d/%d", float64(i)/float64(count)*100, l, i, count)
		}
		raw = parser.ReadLocationRaw(int64(pos))
		// write end ip 4 bytes
		w, _ = wt.Write(iploc.ParseUintIP(end).ReverseBytes())
		n += w
		// write start ip 4 bytes
		indexes = append(indexes, iploc.ParseUintIP(start).ReverseBytes()...)
		// write offset 3 bytes
		indexes = append(indexes, byte(offset), byte(offset>>8), byte(offset>>16))
		// end ip 4 bytes
		offset += 4

		if _, has = fix[end]; has {
			if fix[end][0] != "" {
				if locPos[fix[end][0]] != 0 {
					raw.Text[0] = nil
					raw.Mode[0] = 0x02
					raw.Pos[0] = locPos[fix[end][0]]
				} else {
					raw.Text[0] = []byte(fix[end][0])
					raw.Mode[0] = 0x00
					raw.Pos[0] = offset
				}
			}
			if fix[end][1] != "" {
				if locPos[fix[end][1]] != 0 {
					raw.Text[1] = nil
					raw.Mode[1] = 0x02
					raw.Pos[1] = locPos[fix[end][1]]
				} else {
					raw.Text[1] = []byte(fix[end][0])
					raw.Mode[1] = 0x00
					raw.Pos[1] = uint32(len(raw.Text[0])) + offset
				}
			}
		}

		if raw.Text[0] != nil {
			if raw.Mode[0] != 0x00 {
				w, _ = wt.Write([]byte{raw.Mode[0]})
				n += w
				offset += 1
			}
			locOffset[raw.Pos[0]] = offset
			s = toUTF8(raw.Text[0])
			locPos[string(s)] = raw.Pos[0]
			w, _ = wt.Write(append(s, 0x00))
			n += w
			offset += uint32(len(s)) + 1
		} else {
			o = locOffset[raw.Pos[0]]
			w, _ = wt.Write([]byte{raw.Mode[0], byte(o), byte(o>>8), byte(o>>16)})
			n += w
			offset += 4 // 1+3
		}

		// redirected redundant data (mode 0x02)
		if raw.Text[1] != nil && locOffset[raw.Pos[1]] == 0 {
			locOffset[raw.Pos[1]] = offset
			s = toUTF8(raw.Text[1])
			// CZ88.NET
			if bytes.Compare(s, []byte{32, 67, 90, 56, 56, 46, 78, 69, 84}) == 0 {
				s = []byte{78, 47, 65} // N/A
			}
			locPos[string(s)] = raw.Pos[1]
			w, _ = wt.Write(append(s, 0x00))
			n += w
			offset += uint32(len(s)) + 1
		} else {
			o = locOffset[raw.Pos[1]]
			w, _ = wt.Write([]byte{0x02, byte(o), byte(o>>8), byte(o>>16)})
			n += w
			offset += 4
		}
		return true
	})

	if !quiet {
		fmt.Printf("\rBuild %6.2f%% %*d/%d %s\n", 100.0, l, count, count, time.Since(st))
	}

	min := n + 8
	max := min + len(indexes) - 7
	binary.LittleEndian.PutUint32(head[:4], uint32(min))
	binary.LittleEndian.PutUint32(head[4:], uint32(max))
	wt.Write(indexes)
	wt.Flush()
	f.WriteAt(head[:], io.SeekStart)
	f.Close()

	if !noCheck {
		check()
	}
}

func check() {
	if !quiet {
		fmt.Print("Check ...")
	}
	var (
		locGBK, locUTF8 *iploc.Locator
		err             error
		logger          io.WriteCloser
	)

	if locGBK, err = iploc.Open(qqwrySrc); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	if locUTF8, err = iploc.Open(qqwryDst); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}

	if locGBK.Count() != locUTF8.Count() {
		fmt.Fprintf(os.Stderr, "it's not the same version")
		os.Exit(1)
	}
	if logger, err = os.Create("iploc-conv.log"); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	var (
		detail1, detail2 *iploc.Detail
		count            = locGBK.Count()
		l                = len(fmt.Sprint(count))
		warning          int
		st               = time.Now()
	)
	locGBK.Range(func(i int, start, end iploc.IP) bool {
		detail1 = locGBK.FindIP(start)
		detail2 = locUTF8.FindIP(start)
		if !compare(detail1, detail2, logger) {
			warning ++
		}
		if !quiet {
			fmt.Printf("\rCheck %6.2f%% %*d/%d", float64(i)/float64(count)*100, l, i, count)
		}
		return true
	})
	logger.Close()

	if !quiet {
		fmt.Printf("\rCheck %6.2f%% %*d/%d %s\n", 100.0, l, count, count, time.Since(st))
	}
	if warning > 0 {
		fmt.Printf("%d warnings. please see iploc-conv.log for more details\n", warning)
	} else {
		os.Remove("iploc-conv.log")
	}
}

func compare(a, b *iploc.Detail, logger io.WriteCloser) (ok bool) {
	var s string
	defer func() {
		if !ok {
			if a.Region == " CZ88.NET" && b.Region == "N/A" {
				ok = true
				return
			}
			fmt.Fprintf(logger, " GBK: %s,%s %d,%d\n", a.Start, a.End, a.Start.Uint(), a.End.Uint())
			fmt.Fprintf(logger, "UTF8: %s,%s %d,%d\n", b.Start, b.End, b.Start.Uint(), b.End.Uint())
			if s != "" {
				fmt.Fprintf(logger, " GBK: %s %v\n", s, []byte(s))
				fmt.Fprintf(logger, "UTF8: %s %v\n", b, []byte(b.String()))
				fmt.Fprint(logger, "\n")
			} else {
				fmt.Fprint(logger, "\n")
			}
		}
	}()
	if a.Start.Compare(b.Start) != 0 || a.End.Compare(b.End) != 0 {
		ok = false
		return
	}
	if _, ok = fix[a.End.Uint()]; ok {
		return
	}
	s = string(toUTF8(a.Bytes()))
	ok = s == b.String()
	if s == "" {
		s = "<conversion failure>"
	}
	return
}

func toUTF8(s []byte) (b []byte) {
	r := transform.NewReader(bytes.NewReader(s), gbkDecoder)
	b, _ = ioutil.ReadAll(r)
	return
}
