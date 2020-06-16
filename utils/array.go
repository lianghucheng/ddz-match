package utils

import (
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Shuffle(a []int) []int {
	n := len(a)
	if n == 0 {
		return a
	}
	b := make([]int, n)

	m := rand.Perm(n)
	for i := 0; i < n; i++ {
		b[i] = a[m[i]]
	}
	return b
}

func Index(a []int, sep int) int {
	n := len(a)
	for i := 0; i < n; i++ {
		if a[i] == sep {
			return i
		}
	}
	return -1
}

// 判断 value 是否在 array 中
func InArray(a []int, sep int) bool {
	for _, v := range a {
		if sep == v {
			return true
		}
	}
	return false
}

// 从 array 中移除最开始出现的 sep
func RemoveOnce(a []int, sep int) []int {
	i := Index(a, sep)
	if i == -1 {
		return a
	}
	b := make([]int, 0)
	if i == 0 {
		b = append(b, a[1:]...)
	} else if i == len(a)-1 {
		b = append(b, a[:i]...)
	} else {
		b = append(b, a[:i]...)
		b = append(b, a[i+1:]...)
	}
	return b
}

// 从 array 中移除所有的 sep
func RemoveAll(a []int, sep int) []int {
	b := make([]int, 0)
	for _, v := range a {
		if v != sep {
			b = append(b, v)
		}
	}
	return b
}

func Remove(a []int, sub []int) []int {
	b := append([]int{}, a...)
	for _, v := range sub {
		b = RemoveOnce(b, v)
	}
	return b
}

func ReplaceAll(a []int, old, new int) []int {
	if old == new {
		return a
	}
	if InArray(a, old) {
		b := make([]int, 0)
		for _, v := range a {
			if old == v {
				b = append(b, new)
			} else {
				b = append(b, v)
			}
		}
		return b
	}
	return a
}

func Deduplicate(a []int) []int {
	n := len(a)
	if n == 0 {
		return a
	}
	m := make(map[int]bool)

	b := []int{a[0]}
	m[a[0]] = true
	for i := 1; i < n; i++ {
		if !m[a[i]] {
			b = append(b, a[i])
			m[a[i]] = true
		}
	}
	return b
}

// 比较两个数组的元素是否相等
func Equal(x, y []int) bool {
	if len(x) == len(y) {
		return Contain(x, y)
	}
	return false
}

func Contain(x, y []int) bool {
	if len(x) < len(y) {
		return false
	}
	temp := Deduplicate(y)
	for _, v := range temp {
		if Count(x, v) < Count(y, v) {
			return false
		}
	}
	return true
}

func Count(a []int, sep int) int {
	count := 0
	for _, v := range a {
		if sep == v {
			count++
		}
	}
	return count
}
