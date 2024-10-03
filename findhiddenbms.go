package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func main() {
	start := time.Now()

	flag.Parse()

	path := "./"
	if len(flag.Args()) > 1 {
		fmt.Println("Usage: findhiddenbms <dirpath>")
		os.Exit(1)
	} else if len(flag.Args()) == 1 {
		path = flag.Arg(0)
	}

	fInfo, err := os.Stat(path)
	if err != nil {
		fmt.Println("Wrong path: %w", err)
		os.Exit(1)
	}
	if !fInfo.IsDir() {
		fmt.Println("The entered path is not directory")
		os.Exit(1)
	} else {
		err = findInDirectory(path)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	}

	end := time.Now()
	fmt.Printf("Processing done: %f sec\n", (end.Sub(start)).Seconds())
}

func isBmsPath(path string) bool {
	exts := []string{".bms", ".bme", ".bml", ".pms", ".bmson"}
	return haveExt(path, &exts)
}

func isIgnorePath(path string) bool {
	if isBmsPath(path) {
		return true
	}
	exts := []string{
		".wav", ".ogg", ".mp3", ".bmp", ".png", ".jpg", ".jpeg", ".mpg", ".mpeg", ".wmv", ".avi", ".mp4",
		".ini", ".wc2", ".wos", ".bat", ".db", ".exe", ".mid", ".pdf", ".csv"}
	return haveExt(path, &exts)
}

func isZippedFile(path string) bool {
	exts := []string{".zip", ".lzh", ".rar", ".7z"}
	return haveExt(path, &exts)
}

func isNoCheckFile(path string) bool {
	if isZippedFile(path) {
		return true
	}
	exts := []string{".ini", ".txt", ".db"}
	return haveExt(path, &exts)
}

func haveExt(path string, exts *[]string) bool {
	ext := filepath.Ext(path)
	for _, e := range *exts {
		if strings.ToLower(ext) == e {
			return true
		}
	}
	return false
}

func isIgnoreFileName(name string) bool {
	if isIgnoreChartName(name) {
		return true
	}
	ignoreNames := []string{"read", "lyric", `りどみ`, `りーどみー`, `れあどめ`,
		".ds_store", "テンプレ", "歌詞", "oggenc", "oggdec"}
	return containsName(name, &ignoreNames)
}

func isIgnoreChartName(name string) bool {
	ignoreNames := []string{"blank", "blanc", "noobj", "no_obj", "0_obj", "nokey", "tmp", "temp", "base", "none",
		"empty", "0key", "edit", "hinagata", "haichi", `配置`, `差分`, "定義"}
	return containsName(name, &ignoreNames)
}

func isTargetFileName(name string) bool {
	targetNames := []string{"hidden", "kakushi", "secret", "black"}
	return containsName(name, &targetNames)
}

func containsName(name string, targetNames *[]string) bool {
	for _, n := range *targetNames {
		if strings.Contains(strings.ToLower(name), n) {
			return true
		}
	}
	return false
}

func findInDirectory(dirPath string) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("ReadDir: %w", err)
	}
	for _, f := range files {
		fileName := f.Name()
		filePath := filepath.Join(dirPath, fileName)
		if f.IsDir() {
			err := findInDirectory(filePath)
			if err != nil {
				return err
			}
		} else if !isIgnorePath(fileName) {
			if !isIgnoreFileName(fileName) {
				if isTargetFileName(fileName) {
					fmt.Println("**TargetFilename:", filePath)
				} else if isZippedFile(fileName) {
					fmt.Println("*Zipfile:", filePath)
				}
			}
			if !isIgnoreChartName(fileName) {
				err := readFile(filePath)
				if err != nil {
					return fmt.Errorf("readFile: %w", err)
				}
			}
		} else if !isBmsPath(fileName) {
			if !isNoCheckFile(fileName) {
				isCorrect, err := isCorrectExt(filePath)
				if !isCorrect && err == nil {
					err := readFile(filePath)
					if err != nil {
						return fmt.Errorf("readFile: %w", err)
					}
				}
			} else if filepath.Ext(fileName) == ".txt" {
				err := readFile(filePath)
				if err != nil {
					return fmt.Errorf("readFile: %w", err)
				}
			}
		}
	}
	return nil
}

// extension spoofing check
func isCorrectExt(path string) (bool, error) {
	file, err := os.Open(path) // slow, bottleneck
	if err != nil {
		return false, err
	}
	defer file.Close()

	buf := make([]byte, 4)
	_, err = file.Read(buf)
	if err != nil {
		return false, err
	}

	// file signatures
	signaturesMap := map[string]([]byte){
		".wav":  []byte{0x52, 0x49, 0x46, 0x46},
		".ogg":  []byte{0x4f, 0x67, 0x67, 0x53},
		".mp3":  []byte{0xff}, // {0xff, 0xf3},
		".flac": []byte{0x66, 0x4c, 0x61, 0x43},
		".png":  []byte{0x89, 0x50, 0x4e, 0x47},
		".jpg":  []byte{0xff, 0xd8},
		".jpeg": []byte{0xff, 0xd8},
		".bmp":  []byte{0x42, 0x4d},
		".mp4":  []byte{0x00, 0x00, 0x00},
		".mpg":  []byte{0x00, 0x00, 0x01}, // {0x00, 0x00, 0x01, 0xba}, {0x00, 0x00, 0x01, 0xb3},
		".mpeg": []byte{0x00, 0x00, 0x01},
		".wmv":  []byte{0x30, 0x26, 0xb2, 0x75},
		".avi":  []byte{0x52, 0x49, 0x46, 0x46},
		//".zip": []byte{0x50, 0x4b, 0x03, 0x04},
		".exe": []byte{0x4d, 0x5a},
	}
	ext := strings.ToLower(filepath.Ext(path))
	_, contains := signaturesMap[ext]
	if !contains {
		return false, fmt.Errorf("Not covered by signaturesMap.")
	}
	if bytes.HasPrefix(buf, (signaturesMap[ext])) {
		return true, nil
	}

	return false, nil
}

func readFile(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ReadFile: %w", err)
	}
	if strings.Contains(string(bytes), "*---------------------- HEADER FIELD") {
		isNoObj, err := isNoObjBms(path)
		if err != nil {
			return fmt.Errorf("isNoObjBms: %w", err)
		}
		if !isNoObj {
			fmt.Println("*hasBMS:", path)
		}
	} else if filepath.Ext(path) == ".txt" {
		targetWords := []string{`隠し`, "kakushi", "hidden", "secret", `ひっそり`, `こっそり`}
		for _, word := range targetWords {
			if strings.Contains(string(bytes), word) {
				fmt.Println("**ContainsHiddenWords:", word, path)
				break
			}
		}
	}
	return nil
}

func isNoObjBms(bmspath string) (bool, error) {
	file, err := os.Open(bmspath)
	if err != nil {
		return true, fmt.Errorf("BMSfile open: %w", err)
	}
	defer file.Close()

	const (
		initialBufSize = 10000
		maxBufSize     = 1000000
	)
	scanner := bufio.NewScanner(file)
	buf := make([]byte, initialBufSize)
	scanner.Buffer(buf, maxBufSize)

	for scanner.Scan() {
		line, _, err := transform.String(japanese.ShiftJIS.NewDecoder(), scanner.Text())
		if err != nil {
			return true, fmt.Errorf("ShiftJIS decode: %w", err)
		}

		if regexp.MustCompile(`#[0-9]{5}:.+`).MatchString(line) {
			chint, _ := strconv.Atoi(line[4:6])
			if (chint >= 11 && chint <= 19) || (chint >= 21 && chint <= 29) {
				return false, nil
			}
		}
	}
	if scanner.Err() != nil {
		return true, fmt.Errorf("BMSfile scanner: %w", scanner.Err())
	}
	return true, nil
}
