// Package utils предоставляет утилиты для работы с HMAC и сжатием данных.
package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// CalculateHMAC вычисляет HMAC-SHA256 хеш для данных с использованием секретного ключа.
// Возвращает hex-кодированную строку хеша.
// Параметры:
//   - data: данные для хеширования
//   - key: секретный ключ
func CalculateHMAC(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
