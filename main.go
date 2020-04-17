package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

const version = "0.1.0"

var (
	mode      string
	modeEXIF  = "exif"
	modeMtime = "mtime"

	debug bool
	root  string
	doit  bool
	loc   *time.Location

	existname map[string]bool
)

func main() {
	fh := flag.Bool("h", false, "Show this.")
	fv := flag.Bool("v", false, "Show version.")
	fexif := flag.Bool("exif", false, "Rename with EXIF time.")
	fmtime := flag.Bool("mtime", false, "Rename with modify time.")
	froot := flag.String("root", ".", "File root path.")
	ftz := flag.String("tz", "Asia/Chongqing", "Time zone.")
	fdoit := flag.Bool("doit", false, "Do it (not dry run).")
	fdebug := flag.Bool("debug", false, "Debug mode.")
	flag.Parse()

	if *fh || len(os.Args) <= 1 {
		fmt.Printf("EXAMPLE\n\n" +
			"    re -path ./testdata -exif\n" +
			"    re -path ./testdata -exif -doit\n" +
			"    re -path ./testdata -exif -tz UTC\n\n" +
			"OPTION\n\n")
		flag.PrintDefaults()
		return
	}

	if *fv {
		fmt.Println(version)
		return
	}

	switch {
	case *fexif:
		mode = modeEXIF
	case *fmtime:
		mode = modeMtime
	default:
		fmt.Println("require -exif or -mtime")
		return
	}

	if l, err := time.LoadLocation(*ftz); err != nil {
		log.Fatalf("time zone (%s) error: %v", *ftz, err)
	} else {
		loc = l
	}

	doit = *fdoit
	debug = *fdebug
	existname = make(map[string]bool)

	root = validPath(*froot)
	filepath.Walk(root, rename)
}

func rename(path string, info os.FileInfo, err error) error {
	if err != nil {
		if debug {
			fmt.Println(err)
		}
		return err
	}

	if info.Name()[:1] == "." {
		// ignore all sub files of hidden dir
		if info.IsDir() {
			return filepath.SkipDir
		}
		// ignore hidden file
		return nil
	}

	if info.IsDir() {
		return nil
	}

	tname, err := timename(path, info)
	if err != nil {
		if debug {
			fmt.Printf("%s => %v\n", simplePath(path), err)
		} else {
			fmt.Printf("%s => ?\n", simplePath(path))
		}
		return nil
	}

	newname := fmt.Sprintf("%s_%s", tname, filepath.Base(path))
	newpath := filepath.Join(filepath.Dir(path), newname)

	if existname[newpath] || isExist(newpath) {
		fmt.Printf("%s => [duplication of name]\n", path)
		return nil
	}
	existname[newpath] = true

	fmt.Printf("%s => %s\n", simplePath(path), simplePath(newpath))
	if doit {
		err = os.Rename(path, newpath)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	return nil
}

func simplePath(p string) string {
	if filepath.HasPrefix(p, root) {
		p = p[len(root):]
	}
	if p[:1] == "/" {
		p = p[1:]
	}
	return p
}

func validPath(path string) string {
	if path == "" {
		path = "."
	}

	if p, err := filepath.Abs(path); err != nil {
		log.Fatalf("path(%s) error: %v", path, err)
	} else {
		path = p
	}

	if !isExist(path) {
		log.Fatalf("path(%s) not exist", path)
	}

	return path
}

func timename(path string, info os.FileInfo) (tname string, err error) {
	var t *time.Time
	switch mode {
	case modeEXIF:
		t, err = exiftime(path)
		if err != nil {
			return "", err
		}
	case modeMtime:
		x := info.ModTime()
		t = &x
	default:
		return "", fmt.Errorf("unknown mode: %s", mode)
	}

	return t.In(loc).Format("20060102150405"), nil
}

func exiftime(fname string) (*time.Time, error) {
	f, err := os.Open(fname)
	if err != nil {
		return nil, fmt.Errorf("open file error: %w", err)
	}

	x, err := exif.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode file EXIF error: %w", err)
	}

	t, err := x.DateTime()
	if err != nil {
		return nil, fmt.Errorf("parse file EXIF datetime error: %w", err)
	}

	return &t, nil
}

func isExist(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
