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
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	command        = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	paramSrcDir    = command.String("srcdir", "", "Source directory for comparison")
	paramTargetDir = command.String("targetdir", "", "Target directory for comparison")
	paramProcDir   = command.String("diffdir", "", "Directory with diff files to copy s to the final target")
	paramVerbose   = command.Bool("v", false, "Enable verbose output")
)

type Config struct {
	srcdir    string
	targetdir string
	diffdir   string
	verbose   bool
}

func (config *Config) ParseCommandLine() {
	command.Parse(os.Args[1:])

	if *paramSrcDir == "" {
		fmt.Fprintln(os.Stderr, "Parameter 'srcdir' is empty.")
		command.Usage()
		os.Exit(1)
	}
	if *paramTargetDir == "" {
		fmt.Fprintln(os.Stderr, "Parameter 'targetdir' is empty.")
		command.Usage()
		os.Exit(2)
	}
	if *paramProcDir == "" {
		fmt.Fprintln(os.Stderr, "Parameter 'diffdir' is empty.")
		command.Usage()
		os.Exit(3)
	}

	config.srcdir = *paramSrcDir
	config.targetdir = *paramTargetDir
	config.diffdir = *paramProcDir
	config.verbose = *paramVerbose
}

func (config *Config) copyFiles() error {
	absSrc, err := filepath.Abs(config.srcdir)
	if err != nil {
		return err
	}

	absTarget, err := filepath.Abs(config.targetdir)
	if err != nil {
		return err
	}

	absProc, err := filepath.Abs(config.diffdir)
	if err != nil {
		return err
	}

	dirTarget := filepath.Base(absTarget)
	procTarget := filepath.Join(absProc, dirTarget)

	err = filepath.Walk(absSrc,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			isdir, err := IsDirectory(path)
			if err != nil {
				return err
			}

			newpath := strings.Replace(path, absSrc, absTarget, 1)
			if !isdir {
				if FileExists(newpath) {
					hashSrc, errSrc := GetSha256(path)
					hashTarget, errTarget := GetSha256(newpath)
					if errSrc != nil || errTarget != nil || hashSrc != hashTarget {
						err := filecopy(path, absSrc, procTarget, absProc, config.verbose)
						if err != nil {
							return err
						}
					}
				} else {
					err := filecopy(path, absSrc, procTarget, absProc, config.verbose)
					if err != nil {
						return err
					}
				}
			} else {
				if !FileExists(newpath) {
					if err := os.MkdirAll(newpath, 0777); err != nil {
						return errors.New("failed to create directory: '" + newpath + "', error: '" + err.Error() + "'")
					}
				}
			}
			return nil
		})
	return err
}

func filecopy(path, absSrc, procTarget, absProc string, verbose bool) error {
	copypath := strings.Replace(path, absSrc, procTarget, 1)
	_, err := basecopy(path, copypath)
	if err != nil {
		return errors.New("It was not possible to filecopy file '" + path + "'")
	} else {
		if verbose == true {
			fmt.Println(strings.Replace(path, absSrc, ".", 1), " was copied to ", strings.Replace(copypath, absProc, ".", 1))
		}
	}
	return nil
}

func basecopy(src, dst string) (int64, error) {
	source, err := os.Open(src)
	if err != nil {
		return -1, err
	}
	defer source.Close()

	dstDir := filepath.Dir(dst)
	if !FileExists(dstDir) {
		if err := os.MkdirAll(dstDir, 0777); err != nil {
			return -1, errors.New("failed to create directory: '" + dstDir + "', error: '" + err.Error() + "'")
		}
	}

	destination, err := os.Create(dst)
	if err != nil {
		return -1, err
	}
	defer destination.Close()

	nBytes, err := io.Copy(destination, source)
	if err != nil {
		return -1, err
	}
	return nBytes, nil
}

func FileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		return false
	}
}

func IsDirectory(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), err
}

func GetSha256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func main() {
	config := &Config{}
	config.ParseCommandLine()
	err := config.copyFiles()
	if err != nil {
		fmt.Println("Sync process failed!", err.Error())
		os.Exit(10)
	}
}
