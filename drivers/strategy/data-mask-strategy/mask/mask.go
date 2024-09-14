package mask

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type IMaskDriver interface {
	Exec(body []byte) ([]byte, error)
	String() string
}

type MaskFunc func(origin string) string

var (
	maskByte = '*'
)

const (
	MaskPartialDisplay = "partial-display"
	MaskPartialMask    = "partial-masking"
	MaskTruncation     = "truncation"
	MaskReplacement    = "replacement"
	MaskShuffling      = "shuffling"
)

func GenMaskFunc(cfg *Mask) (MaskFunc, error) {
	switch cfg.Type {
	case MaskPartialDisplay:
		return partialDisplay(cfg.Begin, cfg.Length), nil
	case MaskPartialMask:
		return partialMasking(cfg.Begin, cfg.Length), nil
	case MaskTruncation:
		return truncation(cfg.Begin, cfg.Length), nil
	case MaskReplacement:
		if cfg.Replace == nil {
			return nil, fmt.Errorf("replace is nil")
		}
		return replacement(cfg.Replace.Type, cfg.Replace.Value)
	case MaskShuffling:
		return shuffling(cfg.Begin, cfg.Length), nil
	default:
		return nil, fmt.Errorf("unknown mask type %s", cfg.Type)
	}
}

func partialDisplay(begin int, length int) MaskFunc {
	return func(origin string) string {
		target := strings.Builder{}
		runes := []rune(origin)
		size := len(runes)

		if begin > size {
			for i := 0; i < size; i++ {
				target.WriteRune(maskByte)
			}
		} else if length == -1 || begin+length > size {
			for i := 0; i < begin; i++ {
				target.WriteRune(maskByte)
			}
			for i := begin; i < size; i++ {
				target.WriteRune(runes[i])
			}
		} else {
			for i := 0; i < begin; i++ {
				target.WriteRune(maskByte)
			}
			for i := begin; i < begin+length; i++ {
				target.WriteRune(runes[i])
			}
			for i := begin + length; i < size; i++ {
				target.WriteRune(maskByte)
			}
		}
		return target.String()
	}
}

func partialMasking(begin int, length int) MaskFunc {
	return func(origin string) string {
		target := strings.Builder{}
		runes := []rune(origin)
		size := len(runes)
		if begin > size {
			return origin
		} else if length == -1 || begin+length > size {
			for i := 0; i < begin; i++ {
				target.WriteRune(runes[i])
			}
			for i := begin; i < size; i++ {
				target.WriteRune(maskByte)
			}
		} else {
			for i := 0; i < begin; i++ {
				target.WriteRune(runes[i])
			}
			for i := begin; i < begin+length; i++ {
				target.WriteRune(maskByte)
			}
			for i := begin + length; i < size; i++ {
				target.WriteRune(runes[i])
			}
		}
		return target.String()
	}
}

func truncation(begin int, length int) MaskFunc {
	return func(origin string) string {
		target := strings.Builder{}
		runes := []rune(origin)
		size := len(runes)
		if begin > size {
			return ""
		} else if length == -1 || begin+length > size {
			for i := begin; i < size; i++ {
				target.WriteRune(runes[i])
			}
		} else {
			for i := begin; i < begin+length; i++ {
				target.WriteRune(runes[i])
			}
		}
		return target.String()
	}
}

const (
	ReplaceRandom = "random"
	ReplaceCustom = "custom"
)

func replacement(replaceType string, value string) (MaskFunc, error) {
	switch replaceType {
	case ReplaceRandom:
		return func(origin string) string {
			return replaceWithRandomString(origin)
		}, nil
	case ReplaceCustom:
		return func(origin string) string {
			return value
		}, nil
	default:
		return nil, fmt.Errorf("unknown replace type %s", replaceType)
	}
}

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

// 随机生成指定长度的字符串
func randomString(length int) string {
	// 随机种子
	rand.Seed(time.Now().UnixNano())

	// 创建一个切片，用于存储生成的随机字符
	result := make([]byte, length)

	// 根据长度逐个生成随机字符
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}

	return string(result)
}

// 将原字符串替换为随机生成的小写字母和数字的字符串
func replaceWithRandomString(text string) string {
	// 计算原字符串的长度
	length := len([]rune(text)) // 处理多字节字符
	// 生成与原字符串长度一致的随机字符串
	return randomString(length)
}

func shuffling(begin int, length int) MaskFunc {
	return func(origin string) string {
		runes := []rune(origin)
		// 设置随机数种子，以保证每次运行结果不同
		rand.Seed(time.Now().UnixNano())

		// 使用 Shuffle 函数打乱 nums 切片的顺序
		rand.Shuffle(len(runes), func(i, j int) {
			runes[i], runes[j] = runes[j], runes[i] // 交换切片中第 i 和 j 个元素
		})
		return string(runes)
	}
}
