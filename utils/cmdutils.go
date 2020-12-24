package utils

import "strings"

func SplitArgs(cmd string,delm string)([]string,error){
	return strings.Split(cmd,delm),nil
}