package lib

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func noError(t *testing.T, err error) {
	if err != nil {
		t.Error(err)
	}
}

func equal(t *testing.T, expect, real interface{}) {
	if !reflect.DeepEqual(expect, real) {
		t.Errorf("not equal expect:%v, real:%v", expect, real)
	}
}

func writeData(t *testing.T, dir, fn string, data []byte) {
	fn = filepath.Join(dir, fn)
	var f *os.File
	_, err := os.Stat(fn)
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.Create(fn)
			noError(t, err)
		} else {
			noError(t, err)
		}
	}
	f, err = os.OpenFile(fn, os.O_APPEND|os.O_RDWR, 0644)
	noError(t, err)
	defer f.Close()
	_, err = f.Write(data)
	noError(t, err)
}

func TestTailDir(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	noError(t, err)
	log.Printf("test dir %v", dir)
	defer os.RemoveAll(dir)
	writeData(t, dir, "f", []byte("hi"))
	r, err := NewDirReader(Config{Dir: dir})
	noError(t, err)
	p := make([]byte, 2)
	// file d
	{
		data := []byte("hi")
		go writeData(t, dir, "d", data)
		n, err := io.ReadFull(r, p)
		noError(t, err)
		equal(t, 2, n)
		equal(t, data, p)
	}
	// file d append
	{
		data := []byte("ab")
		go writeData(t, dir, "d", data)
		n, err := io.ReadFull(r, p)
		noError(t, err)
		equal(t, 2, n)
		equal(t, data, p)
	}
	// new file a
	{
		data := []byte("df")
		go writeData(t, dir, "a", data)
		n, err := io.ReadFull(r, p)
		noError(t, err)
		equal(t, 2, n)
		equal(t, data, p)
	}
}
