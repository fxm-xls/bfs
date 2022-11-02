/**
 * @Author:
 * @File: util
 * @Version: 1.0.0
 * @Date: 2021/7/28 15:02
 * @Description:
 */
package utils

import (
	"fmt"
	"github.com/google/uuid"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	notFormal = []string{"\\", "$", "(", ")", "*", "+", ".", "[", "]", "?", "^", "|", "{", "}"}
)

func IsContainsInt(items []int, item int) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

func IsContainsStr(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// s2包含 sub1 slice .
func Subslice(sub1 []string, s2 []string) bool {
	if len(sub1) > len(s2) {
		return false
	}
	for _, e := range sub1 {
		if !IsContainsStr(s2, e) {
			return false
		}
	}
	return true
}

func Intersect(slice1 []string, slice2 []string) []string { //交集
	m := make(map[string]int)
	n := make([]string, 0)
	for _, v := range slice1 {
		m[v]++
	}
	for _, v := range slice2 {
		times, _ := m[v]
		if times == 1 {
			n = append(n, v)
		}
	}
	return n
}

func SliceInt2str(input []int) (res []string) {
	for _, id := range input {
		res = append(res, fmt.Sprint(id))
	}
	return
}

func SliceStr2Int(input []string) (res []int) {
	for _, id := range input {
		res = append(res, Str2Int(id))
	}
	return
}

func Difference(slice1, slice2 []string) []string { //全集-交集
	m := make(map[string]int)
	n := make([]string, 0)
	inter := Intersect(slice1, slice2)
	for _, v := range inter {
		m[v]++
	}
	for _, value := range slice1 {
		if m[value] == 0 {
			n = append(n, value)
		}
	}

	for _, v := range slice2 {
		if m[v] == 0 {
			n = append(n, v)
		}
	}
	return n
}

func RegexReplace(in string) string {
	for _, one := range notFormal {
		in = strings.ReplaceAll(in, one, "\\"+one)
	}
	return in
}

func Str2Int(in string) int {
	tmp, _ := strconv.Atoi(in)
	return tmp
}

// MustUUID 创建UUID，如果发生错误则抛出panic
func MustUUID() string {
	v, err := NewUUID()
	if err != nil {
		panic(err)
	}
	return v
}

// NewUUID 创建UUID
func NewUUID() (string, error) {
	v, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return v.String(), nil
}

// 两个数组取并集
func UnionSlice(nums1 []string, nums2 []string) []string {
	m := make(map[string]int)
	for _, v := range nums1 {
		m[v]++
	}
	for _, v := range nums2 {
		times, _ := m[v]
		if times == 0 {
			nums1 = append(nums1, v)
		}
	}
	return nums1
}

// 判断目录是否存在
func IsExistDir(filePath string) bool {
	s, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	if s == nil {
		return false
	} else {
		return s.IsDir()
	}
}

// 判断文件是否存在
func IsExistFile(file string) bool {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

// GetStrLength 返回输入的字符串的字数，汉字、中文标点、英文和其他字符都算 1 个字数
func GetStrLength(str string) float64 {
	var total float64

	reg := regexp.MustCompile("/·|，|。|《|》|‘|’|”|“|；|：|【|】|？|（|）|、/")

	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) || reg.Match([]byte(string(r))) {
			total = total + 1
		} else {
			total = total + 1
		}
	}

	return math.Ceil(total)
}

func DelItem(vs []string, s string) []string {
	for i := 0; i < len(vs); i++ {
		if s == vs[i] {
			vs = append(vs[:i], vs[i+1:]...)
			i = i - 1
		}
	}
	return vs
}

func DelItems(vs []string, dels []string) []string {
	dMap := make(map[string]bool)
	for _, s := range dels {
		dMap[s] = true
	}

	for i := 0; i < len(vs); i++ {
		if _, ok := dMap[vs[i]]; ok {
			vs = append(vs[:i], vs[i+1:]...)
			i = i - 1
		}
	}
	return vs
}
