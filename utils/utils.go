package utils

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func Authenticate(r *http.Request) (string, uint, error) {
	tokenCookie, err := r.Cookie("token")
	if err != nil {
		return "", 0, errors.New("no token")
	}
	token := tokenCookie.Value

	groupNumberCookie, err := r.Cookie("group_number")
	if err != nil {
		return "", 0, errors.New("no group_number")
	}
	groupNumber, err := strconv.Atoi(groupNumberCookie.Value)
	if err != nil {
		return "", 0, errors.New("invalid group number")
	}

	return token, uint(groupNumber), nil
}

func JoinUintSlice(slice []uint, sep string) string {
	var builder strings.Builder

	for i, value := range slice {
		if i > 0 {
			// 在第一个元素之后的元素前添加分隔符
			builder.WriteString(sep)
		}
		builder.WriteString(fmt.Sprint(value))
	}

	return builder.String()
}
