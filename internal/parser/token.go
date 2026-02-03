package parser

import (
	"github.com/mattn/go-shellwords"
)

// Tokenize 使用 shellwords 对命令进行分词
func Tokenize(cmd string) ([]string, error) {
	return shellwords.Parse(cmd)
}
