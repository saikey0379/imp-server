package utils

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func FileExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}

func GetDir(path string) (dir string, file string) {
	dr := strings.Split(path, "/")
	var dirArr []string
	for i, v := range dr {
		if v != "" && i != len(dr)-1 {
			dirArr = append(dirArr, v)
		}
		if i == len(dr)-1 {
			file = v
		}
	}
	return fmt.Sprintf("/%s", strings.Join(dirArr, "/")), file
}

func IsFileMode(modString string) bool {
	_, err := strconv.ParseUint(modString, 8, 32)
	if err != nil {
		return false
	}
	return true
}

func ConvertToFileMode(modString string) (fileMod os.FileMode, ok bool) {
	modeValue, err := strconv.ParseUint(modString, 8, 32)
	if err != nil {
		return fileMod, false
	}

	// 将无符号整数转换为 os.FileMode 类型
	fileMod = os.FileMode(modeValue)
	return fileMod, true
}
