package utils

import (
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

// 分组字符串
// 例如：输入[ "Hong Kong 01", "Hong Kong 02", "USA 1", "Taiwan 1", "USA 2", "Taiwan 2"]
// 输出 ["Hong Kong 01", "Hong Kong 02"] ["USA 1","USA 2"] ["Taiwan 1","Taiwan 2"]
func GroupStringArray(strs []string, prefixes []string) [][]string {
	strsLen := len(strs)
	switch strsLen {
	case 0:
		return nil
	case 1:
		return [][]string{{strs[0]}}
	}

	var groups [][]string

	strsMap := make(map[string]struct{})

	for i := range strs {
		strsMap[strs[i]] = struct{}{}
	}
	// 遍历prefixes，把strs数据中元素先查找prefixes中元素，如果匹配，则预分组，然后把不包含prefixes中元素的，继续分组
	// 假如prefixes为 ["Hong Kong","Taiwan"]，strs为["Hong Kong 01","Hong Kong 02", "Taiwan 01", "Twaiwan 02", "USA","Japan"]
	// 那么先把他分成["Hong Kong 01", "Hong Kong 02"] ["Taiwan 01", "Taiwan 02"] ["USA", "Japan"]
	// 前两个数组直接插入结果数组中，后面的数组继续往下处理
	for j := range prefixes {
		var tmpGroup []string
		for key := range strsMap {
			if strings.Contains(key, prefixes[j]) {
				tmpGroup = append(tmpGroup, key)
				// remove key
				delete(strsMap, key)
			}
		}
		// add tmpGroup to result array
		groups = append(groups, tmpGroup)
		logrus.Infof("append %v to result\n", tmpGroup)
	}

	logrus.Infof("len strs: %v\n", len(strs))
	// replace strs with strsMap's keys
	newStrs := []string{}
	for key := range strsMap {
		newStrs = append(newStrs, key)
	}
	logrus.Infof("len newStrs: %v\n", len(newStrs))

	sort.Strings(newStrs)
	logrus.Debugf("begin: %v\n", newStrs)
	// left,right定义了一个区间[left,right]，分组过程如下：
	// left保持不变，right右移，检查区间中最长公共前缀，公共前缀突然减小了，
	// 那么扔掉当前right，此时区间[left,right)中的元素是同一个分组的，
	// 然后left取值right，right取值right+1,继续查找
	var (
		left    = 0
		right   = 1
		lcp     = ""
		lastLcp = ""
	)
	strsLen = len(newStrs)
	for right < strsLen {
		var tmpStrs []string
		for i := left; i <= right; i++ {
			tmpStrs = append(tmpStrs, newStrs[i])
		}
		lcp = LongestCommonPrefix(tmpStrs)
		logrus.Debugf("lcp of %v is %v\n", tmpStrs, lcp)
		if lcp == "" || len(lcp) < len(lastLcp) {
			// found!
			// 注意：找到的时候，不包含right指向的元素，因为就是它导致了lcp变短了
			logrus.Debugf("len small find\n")
			lastLcp = ""
			groups = append(groups, tmpStrs[:len(tmpStrs)-1])
			left = right
			right = left + 1
			if right == strsLen {
				// 如果最后只剩一个元素
				logrus.Debugf("find only one\n")
				groups = append(groups, []string{newStrs[left]})
				break
			}
			continue
		}
		if right == strsLen-1 {
			logrus.Debugf("find last\n")
			groups = append(groups, tmpStrs)
			break
		}
		lastLcp = lcp
		right += 1
	}
	return groups
}
