package base

import ()

var password_table = "abcdefghijklmnopqrstuvwxyz0123456789!@#$*ABCDEFGHIJKLMNOPQRSTUVWXYZ"

const DEF_PASSWORD_LEN = 16

func Password(length ...int) string {

	count := DEF_PASSWORD_LEN

	if len(length) > 0 && length[0] > 0 {
		count = length[0]
	}

	str := make([]byte, 0, count)

	for i := 0; i < count; i++ {
		str = append(str, password_table[RandBetween(0, int32(len(password_table)-1))])
	}

	return string(str)
}

var assic_table = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandString(length int) string {

	str := make([]byte, 0, length)

	for i := 0; i < length; i++ {
		str = append(str, assic_table[RandBetween(0, int32(len(assic_table)-1))])
	}

	return string(str)
}
