/*
 * Copyright 2020 Intershop Communications AG.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a filecopy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

var (
	testdirname = "testprojectdir"
	dirs        = [2]string{"src", "target"}
	subdirs     = [4]string{"apps", "cluster", "domains", "servletEngine"}
)

func GetTestDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	testdir := filepath.Join(dir, testdirname)
	return testdir
}

func CreateFile(path string, bytearr []byte) error {
	return ioutil.WriteFile(path, bytearr, 0644)
}

func setupTestCase(t *testing.T) func(t *testing.T) {
	t.Log("setup test case")

	testdirpath := GetTestDir()
	os.MkdirAll(testdirpath, os.ModePerm)

	for _, d := range dirs {
		confdir := filepath.Join(testdirpath, d, "system-conf")

		for _, sd := range subdirs {
			os.MkdirAll(filepath.Join(confdir, sd), os.ModePerm)
		}
	}

	return func(t *testing.T) {
		t.Log("teardown test case")

		os.RemoveAll(testdirpath)
	}
}

func TestParseCommand(t *testing.T) {
	os.Args = []string{"command", "--srcdir=/path/to/srcdir", "--targetdir=/path/to/targetdir", "--diffdir=/path/to/diffdir"}
	m := &Config{}
	m.ParseCommandLine()

	if m.srcdir != "/path/to/srcdir" {
		t.Errorf("Src-directory setting is not ok. It is %s and should be %s", m.srcdir, "/path/to/srcdir")
	}
	if m.targetdir != "/path/to/targetdir" {
		t.Errorf("Target-directory setting is not ok. It is %s and should be %s", m.targetdir, "/path/to/targetdir")
	}
	if m.diffdir != "/path/to/diffdir" {
		t.Errorf("Diff-directory setting is not ok. It is %s and should be %s", m.diffdir, "/path/to/diffdir")
	}
}

func setupChangedFile() {
	testdirpath := GetTestDir()
	testfileSrc := filepath.Join(testdirpath, dirs[0], "system-conf", subdirs[1], "testfile.properties")
	testfileTarget := filepath.Join(testdirpath, dirs[1], "system-conf", subdirs[1], "testfile.properties")

	testdirSrc := filepath.Join(testdirpath, dirs[0], "system-conf", subdirs[1], "testdir")
	os.MkdirAll(testdirSrc, os.ModePerm)

	CreateFile(testfileSrc, []byte("hello\ntest1\n"))
	CreateFile(testfileTarget, []byte("hello\ntest2\n"))
}

func TestChangedFile(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	setupChangedFile()
	testdirpath := GetTestDir()

	procPath := filepath.Join(testdirpath, "proc")

	m := &Config{}
	m.srcdir = filepath.Join(testdirpath, dirs[0], "system-conf")
	m.targetdir = filepath.Join(testdirpath, dirs[1], "system-conf")
	m.diffdir = procPath

	m.copyFiles()

	procFile := filepath.Join(procPath, "system-conf", subdirs[1], "testfile.properties")
	procDir := filepath.Join(testdirpath, dirs[1], "system-conf", subdirs[1], "testdir")

	if FileExists(procFile) {
		dat, err := ioutil.ReadFile(procFile)
		if err != nil {
			t.Fatalf("File " + procFile + " is not readable.")
		}
		if string(dat) != "hello\ntest1\n" {
			t.Fatalf("File " + procFile + " has wrong content.")
		}
	} else {
		t.Fatalf("File " + procFile + " does not exists.")
	}
	isdir, _ := IsDirectory(procDir)
	if !FileExists(procDir) || !isdir {
		t.Fatalf("Directory " + procDir + " does not exists.")
	}
}

func setupChangedFiles() {
	testdirpath := GetTestDir()

	testfileSrc1 := filepath.Join(testdirpath, dirs[0], "system-conf", subdirs[1], "testfile1.properties")
	CreateFile(testfileSrc1, []byte("hello\ntestfile 1\ntest1\n"))

	testfileSrc2 := filepath.Join(testdirpath, dirs[0], "system-conf", subdirs[1], "testfile2.properties")
	CreateFile(testfileSrc2, []byte("hello\ntestfile 2\ntest1\n"))

	testfileTarget := filepath.Join(testdirpath, dirs[1], "system-conf", subdirs[1], "testfile.properties")
	CreateFile(testfileTarget, []byte("hello\ntest2\n"))
}

func TestChangedFiles(t *testing.T) {
	teardownTestCase := setupTestCase(t)
	defer teardownTestCase(t)

	setupChangedFiles()
	testdirpath := GetTestDir()

	procPath := filepath.Join(testdirpath, "proc")

	m := &Config{}
	m.srcdir = filepath.Join(testdirpath, dirs[0], "system-conf")
	m.targetdir = filepath.Join(testdirpath, dirs[1], "system-conf")
	m.diffdir = procPath

	m.copyFiles()

	procFile1 := filepath.Join(procPath, "system-conf", subdirs[1], "testfile1.properties")
	if FileExists(procFile1) {
		dat, err := ioutil.ReadFile(procFile1)
		if err != nil {
			t.Fatalf("File " + procFile1 + " is not readable.")
		}
		if string(dat) != "hello\ntestfile 1\ntest1\n" {
			t.Fatalf("File " + procFile1 + " has wrong content.")
		}
	} else {
		t.Fatalf("File " + procFile1 + " does not exists.")
	}
	procFile2 := filepath.Join(procPath, "system-conf", subdirs[1], "testfile2.properties")
	if !FileExists(procFile2) {
		t.Fatalf("File " + procFile2 + " does not exists.")
	}
}
