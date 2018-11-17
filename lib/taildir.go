package lib

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

func (d *DirReader) getNextFiles(currentFile string) (file string, err error) {
	fileDirs, err := ioutil.ReadDir(d.conf.Dir)
	if err != nil {
		log.Printf("ReadDir failed %v %v", err, d.conf.Dir)
		return
	}
	var files []os.FileInfo
	for _, fd := range fileDirs {
		matched, _ := filepath.Match(d.conf.FilePattern, fd.Name())
		if !fd.IsDir() && matched {
			files = append(files, fd)
		}
	}
	if len(files) == 0 {
		return
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})
	if currentFile == "" {
		return filepath.Join(d.conf.Dir, files[0].Name()), nil
	}
	fi, err := os.Stat(currentFile)
	if err != nil {
		log.Printf("Stat file failed %v %v", err, currentFile)
		return
	}
	for i := len(files) - 1; i >= 0; i-- {
		if fi.ModTime().Before(files[i].ModTime()) {
			return filepath.Join(d.conf.Dir, files[i].Name()), nil
		}
	}
	return
}

type DirReader struct {
	conf Config
	f    *os.File
	once sync.Once
}

func (d *DirReader) Read(b []byte) (n int, err error) {
	d.once.Do(func() { err = d.openFirstFile() })
	if err != nil {
		return
	}
	n, err = d.f.Read(b)
	if err == nil || err != io.EOF {
		return
	}
	next, err := d.getNextFiles(d.f.Name())
	if err != nil {
		return
	}
	if next == "" {
		err = nil
		time.Sleep(time.Millisecond * 100)
		return
	}
	n1, err := d.f.Read(b[n:])
	if err == nil || err != io.EOF {
		return n + n1, err
	}
	d.f.Close()
	log.Printf("close: %v and open %v", d.f.Name(), next)
	d.f, err = os.Open(next)
	if err != nil {
		log.Printf("Open file failed %v %v", err, next)
		return
	}
	return
}

func (d *DirReader) openFirstFile() (err error) {
	for {
		var firstFileInDir bool
		fn, err := d.getNextFiles("")
		if err != nil {
			return err
		}
		if fn == "" {
			firstFileInDir = true
			time.Sleep(time.Millisecond * 100)
			continue
		}
		f, err := os.Open(fn)
		if err != nil {
			log.Printf("%v, fn:%v", err, fn)
			return err
		}
		var offset int64
		if !firstFileInDir {
			offset, err = f.Seek(0, os.SEEK_END)
			if err != nil {
				log.Printf("%v, fn:%v", err, fn)
				return err
			}
		}
		log.Printf("start from file %v:%v", fn, offset)
		d.f = f
		break
	}
	return
}

type Config struct {
	Dir         string
	FilePattern string
}

func NewDirReader(conf Config) (io.Reader, error) {
	if conf.Dir == "" {
		conf.Dir = "."
	}
	fi, err := os.Stat(conf.Dir)
	if err != nil {
		log.Printf("read dir: %v failed: %v", conf.Dir, err)
		return nil, err
	}
	if !fi.IsDir() {
		err = fmt.Errorf("not a dir: %v", conf.Dir)
		log.Printf("%v", err)
		return nil, err
	}
	if conf.FilePattern == "" {
		conf.FilePattern = "*"
	}
	if _, err = filepath.Match(conf.FilePattern, ""); err != nil {
		log.Printf("invalid pattern %v %v", conf.FilePattern, err)
		return nil, err
	}
	return &DirReader{
		conf: conf,
	}, nil
}
