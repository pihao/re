package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

const version = "1.1.0"

const (
	modeExif  = "exif"
	modeMtime = "mtime"
)

var (
	loc *time.Location

	root      string
	existname map[string]bool
)

var option = struct {
	showHelp    bool
	showVersion bool
	debug       bool
	dryRun      bool
	path        string
	mode        string
	timezone    string
	recursive   bool
}{}

func main() {
	flag.BoolVar(&option.showHelp, "h", false, "Show this.")
	flag.BoolVar(&option.showVersion, "v", false, "Show version.")
	flag.BoolVar(&option.debug, "d", false, "Debug mode.")
	flag.BoolVar(&option.dryRun, "n", false, "Do it (not dry run).")
	flag.StringVar(&option.path, "p", ".", "File path.")
	flag.StringVar(&option.mode, "m", "exif", "Rename mode. valid option: exif, mtime.")
	flag.StringVar(&option.timezone, "t", "Asia/Chongqing", "Time zone.")
	flag.BoolVar(&option.recursive, "s", true, "recursive into directories.")
	flag.Parse()

	if option.showHelp || len(os.Args) <= 1 {
		fmt.Printf("EXAMPLE\n\n" +
			"    re -p ./testdata -e -n\n" +
			"    re -p ./testdata -e\n" +
			"    re -p ./testdata -e -t UTC\n\n" +
			"OPTION\n\n")
		flag.PrintDefaults()
		return
	}

	if option.showVersion {
		fmt.Println(version)
		return
	}

	if slices.Contains([]string{modeExif, modeMtime}, option.mode) {
		fmt.Println("require one of -e and -m option")
		return
	}

	if l, err := time.LoadLocation(option.timezone); err != nil {
		log.Fatalf("time zone (%s) error: %v", option.timezone, err)
	} else {
		loc = l
	}

	existname = make(map[string]bool)
	root = validPath(option.path)
	filepath.Walk(root, rename)
}

func rename(path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Println(err)
		return nil
	}

	if path == root {
		if option.debug {
			fmt.Printf("ignore: root path: %s\n", relPath(path))
		}
		return nil
	}

	if !option.recursive && info.IsDir() && len(filepath.Dir(path)) >= len(root) {
		if option.debug {
			fmt.Printf("ignore: sub-directory: %s\n", relPath(path))
		}
		return filepath.SkipDir
	}

	if filepath.Dir(path) != root {
		if option.debug {
			fmt.Printf("ignore: sub-directory file: %s\n", relPath(path))
		}
		return nil
	}

	if info.Name()[:1] == "." {
		if option.debug {
			fmt.Printf("ignore: hidden file: %s\n", relPath(path))
		}
		return nil
	}

	tname, err := timename(path, info)
	if err != nil {
		if option.debug {
			fmt.Printf("%s => %v\n", relPath(path), err)
		} else {
			fmt.Printf("%s => ?\n", relPath(path))
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

	fmt.Printf("%s => %s\n", relPath(path), relPath(newpath))

	if !option.dryRun {
		err = os.Rename(path, newpath)
		if err != nil {
			fmt.Println(err)
			return nil
		}
	}

	return nil
}

func relPath(p string) string {
	r, _ := filepath.Rel(root, p)
	return r
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
	switch option.mode {
	case modeExif:
		t, err = exiftime(path)
		if err != nil {
			return "", err
		}
	case modeMtime:
		x := info.ModTime()
		t = &x
	default:
		return "", fmt.Errorf("invalid mode: %s", option.mode)
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
	return os.IsExist(err)
}
