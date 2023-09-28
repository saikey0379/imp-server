package utils

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type DiffRst struct {
	Linea   string
	Lineb   string
	Sign    string
	Content string
}

// PingLoop return when success
func PingLoop(host string, pkgCnt int, timeout int) bool {
	for i := 0; i < pkgCnt; i++ {
		if Ping(host, timeout) {
			return true
		}
	}
	return false
}

func IsIpAddress(host string) bool {
	var add = strings.Split(host, ".")
	if len(add) != 4 {
		return false
	}

	for i, j := range add {
		jint, err := strconv.Atoi(j)
		if err != nil {
			return false
		}
		if i == 0 || i == 3 {
			if jint <= 0 || jint >= 255 {
				return false
			}
		}

		if jint < 0 || jint > 255 {
			return false
		}
	}

	return true
}

func GetMd5ByFile(file string) (result string, err error) {
	byte_file, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	md5sum := md5.Sum(byte_file)
	result = hex.EncodeToString(md5sum[:])
	return result, nil
}

func FileDiff(filea string, fileb string) (result []DiffRst, err error) {
	var filea_strs, fileb_strs []string
	filea_bytes, err := ioutil.ReadFile(filea)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(filea_bytes), "\n") {
		filea_strs = append(filea_strs, line)
	}

	fileb_bytes, err := ioutil.ReadFile(fileb)
	if err != nil {
		return nil, err
	}
	for _, line := range strings.Split(string(fileb_bytes), "\n") {
		fileb_strs = append(fileb_strs, line)
	}

	filea_rm := GetDuplicateStr(filea_strs)
	fileb_rm := GetDuplicateStr(fileb_strs)

	var diffrst, diff_filea, diff_fileb []DiffRst
	var diff DiffRst
	var filebidx = 0
	//Get diffrst diff_filea
	for x, i := range filea_strs {
		match := false
		for z, j := range fileb_strs {
			if i == j && i != "" && !IsValueInList(i, filea_rm) && !IsValueInList(i, fileb_rm) {
				if z >= filebidx {
					match = true
					filebidx = z
					diff.Linea = fmt.Sprintf("%-3s", strconv.Itoa(x+1))
					diff.Lineb = fmt.Sprintf("%-3s", strconv.Itoa(z+1))
					diff.Sign = ""
					diff.Content = j
					diffrst = append(diffrst, diff)
					break
				}
			}
		}

		if match == false && x < len(filea_strs)-1 {
			diff.Linea = fmt.Sprintf("%-3s", strconv.Itoa(x+1))
			diff.Lineb = fmt.Sprintf("%-3s", "")
			diff.Content = i
			diff.Sign = "-"
			diff_filea = append(diff_filea, diff)
		}

	}
	//Get diff_fileb
	for x, i := range fileb_strs {
		match := false
		for _, j := range diffrst {
			if i == j.Content {
				match = true
				break
			}
		}
		if match == false && x < len(fileb_strs)-1 {
			diff.Linea = fmt.Sprintf("%-3s", "")
			diff.Lineb = fmt.Sprintf("%-3s", strconv.Itoa(x+1))
			diff.Content = i
			diff.Sign = "+"
			diff_fileb = append(diff_fileb, diff)
		}

	}

	//add diffrst diff_filea

	for _, i := range diff_filea {
		var match = false
		linea_i, err := strconv.Atoi(strings.TrimSpace(i.Linea))
		if err != nil {
			fmt.Println(err.Error())
		}

		for k, j := range diffrst {
			linea_j, err := strconv.Atoi(strings.TrimSpace(j.Linea))
			if err != nil {
				fmt.Println(err.Error())
			}

			if linea_i < linea_j {
				match = true
				diffrst = append(diffrst[:k+1], diffrst[k:]...)
				diffrst[k] = i
				break
			}
		}

		if !match {
			diffrst = append(diffrst, i)
		}
	}
	//add diffrst diff_fileb
	for _, i := range diff_fileb {
		var match = false
		lineb_i, err := strconv.Atoi(strings.TrimSpace(i.Lineb))
		if err != nil {
			fmt.Println(err.Error())
		}

		for k, j := range diffrst {
			var lineb_j int
			if strings.TrimSpace(j.Lineb) != "" {
				lineb_j, err = strconv.Atoi(strings.TrimSpace(j.Lineb))
				if err != nil {
					fmt.Println(err.Error())
				}
			}

			if lineb_i < lineb_j {
				match = true
				diffrst = append(diffrst[:k+1], diffrst[k:]...)
				diffrst[k] = i
				break
			}

		}

		if !match {
			diffrst = append(diffrst, i)
		}
	}

	//合并"-/+",排序
	for v, i := range diffrst {
		var match_index = 0
		var match_linea = 0
		if strings.TrimSpace(i.Lineb) == "" {
			for j := v + 1; j < len(diffrst); j++ {
				if match_index != 0 {
					break
				}

				if strings.TrimSpace(diffrst[j].Linea) != "" && strings.TrimSpace(diffrst[j].Lineb) != "" {
					match_index = j
				} else if j == len(diffrst)-1 {
					match_index = j + 1
				}

			}

			for j := v + 1; j < len(diffrst); j++ {
				if strings.TrimSpace(diffrst[j].Linea) == "" && diffrst[j].Content == i.Content && j <= match_index {
					match_linea = j
					diffrst[v].Lineb = diffrst[j].Lineb
					diffrst[v].Sign = ""
					diffrst = append(diffrst[:j], diffrst[j+1:]...)
					break

				}
			}

			if match_linea > 0 {
				for j := match_linea - 1; j > v; j-- {
					if strings.TrimSpace(diffrst[match_linea-1].Sign) == "+" {
						diffrst = append(diffrst[:v+1], diffrst[v:]...)
						diffrst[v] = diffrst[match_linea]
						diffrst = append(diffrst[:match_linea], diffrst[match_linea+1:]...)
					}
				}
			}

		}

	}
	return diffrst, nil
}

func GetDuplicateStr(arr []string) (newArr []string) {
	newArr = make([]string, 0)
	for i := 0; i < len(arr); i++ {
		repeat := false
		for j := i + 1; j < len(arr); j++ {
			if arr[i] == arr[j] {
				repeat = true
				break
			}
		}
		if repeat {
			newArr = append(newArr, arr[i])
		}
	}
	return
}

func IsValueInList(value string, list []string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}

func WriteFile(filepath string, content string, mod int) error {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(mod))
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Error opening file:[%s]", err.Error()))
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(content)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("Error writing file:[%s]", err.Error()))
	}
	writer.Flush()
	return nil
}

func IsInArrayUint(ui uint, arrUi []uint) bool {
	for _, value := range arrUi {
		if ui == value {
			return true
		}
	}
	return false
}
