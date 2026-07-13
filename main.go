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

const version = "1.2.0"

const (
	modeExif  = "exif"  // EXIF time.
	modeMtime = "mtime" // Modify time.
	modeBtime = "btime" // Birth/Create time.
)

var (
	loc *time.Location

	root string
)

var option = struct {
	showHelp    bool
	showVersion bool
	debug       bool
	dryRun      bool
	path        string
	mode        string
	timezone    string
}{}

func main() {
	flag.BoolVar(&option.showHelp, "h", false, "Show this.")
	flag.BoolVar(&option.showVersion, "v", false, "Show version.")
	flag.BoolVar(&option.debug, "d", false, "Debug mode.")
	flag.BoolVar(&option.dryRun, "n", false, "Dry run.")
	flag.StringVar(&option.path, "p", ".", "File path.")
	flag.StringVar(&option.mode, "m", "exif", "Rename mode. valid option: exif, mtime, btime.")
	flag.StringVar(&option.timezone, "t", "Asia/Chongqing", "Time zone.")
	flag.Parse()

	if option.showHelp || len(os.Args) <= 1 {
		fmt.Printf("EXAMPLE\n\n" +
			"    re -p ./testdata -m exif -n\n" +
			"    re -p ./testdata -m exif\n" +
			"    re -p ./testdata -m exif -t UTC\n\n" +
			"OPTION\n\n")
		flag.PrintDefaults()
		return
	}

	if option.showVersion {
		fmt.Println(version)
		return
	}

	if !slices.Contains([]string{modeExif, modeMtime, modeBtime}, option.mode) {
		fmt.Println("require one of -e and -m option")
		return
	}

	if l, err := time.LoadLocation(option.timezone); err != nil {
		log.Fatalf("time zone (%s) error: %v", option.timezone, err)
	} else {
		loc = l
	}

	root = mustFormatPath(option.path)
	if err := filepath.Walk(root, collect); err != nil {
		log.Fatal(err)
	}
	rename(files)
}

var (
	existname = map[string]bool{}
	files     = [][2]string{}
)

func rename(files [][2]string) {
	for _, e := range files {
		fmt.Printf("%s => %s\n", relPath(e[0]), relPath(e[1]))
		if !option.dryRun {
			if err := os.Rename(e[0], e[1]); err != nil {
				fmt.Println(err)
			}
		}
	}
}

func collect(path string, info os.FileInfo, err error) error {
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

	if len(filepath.Dir(path)) > len(root) {
		if option.debug {
			fmt.Printf("ignore: recursive: %s\n", relPath(path))
		}
		if info.IsDir() {
			return filepath.SkipDir
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

	if existname[newpath] || !isNotExist(newpath) {
		fmt.Printf("%s => [duplication of name]\n", path)
		return nil
	}
	existname[newpath] = true

	files = append(files, [2]string{path, newpath})

	return nil
}

func relPath(p string) string {
	r, _ := filepath.Rel(root, p)
	return r
}

func mustFormatPath(path string) string {
	if path == "" {
		path = "."
	}

	if p, err := filepath.Abs(path); err != nil {
		log.Fatalf("path(%s) error: %v", path, err)
	} else {
		path = p
	}

	if isNotExist(path) {
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
	case modeBtime:
		x := GetBirthTime(info)
		t = &x
	default:
		return "", fmt.Errorf("invalid mode: %s", option.mode)
	}

	if t.IsZero() {
		return "", fmt.Errorf("failed to get time")
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

func isNotExist(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}
