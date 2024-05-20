package decode

const generator = 0b1011

// Функция для декодирования 7-битного блока
func blockDecode(block byte) (byte, bool) {
	originalBlock := block
	// Выполняем деление блоков, чтобы получить синдром (контрольные биты)
	for i := 6; i >= 3; i-- {
		if (block>>i)&1 != 0 {
			block ^= generator << (i - 3)
		}
	}

	// Если синдром не нулевой, значит есть ошибка
	if block != 0 {
		// Ищем и исправляем ошибку
		errorPosition := int(block & 0x07)
		originalBlock ^= 1 << errorPosition
		return originalBlock >> 3, true
	}

	// Возвращаем декодированный блок
	return originalBlock >> 3, false
}

// DataDecode Функция для декодирования всех данных
func DataDecode(data []byte) ([]byte, int) {
	var decodedData []byte
	var totalErrors int
	var buffer byte
	var hasBuffer = false

	for _, b := range data {
		// Декодируем блок
		decodedBlock, hasError := blockDecode(b)
		if hasError {
			totalErrors++
		}

		if !hasBuffer {
			// Сохраняем первый блок в буфер
			buffer = decodedBlock
			hasBuffer = true
		} else {
			// Объединяем два блока по 4 бита в один байт
			combinedBlock := (buffer << 4) | (decodedBlock & 0x0F)
			decodedData = append(decodedData, combinedBlock)
			hasBuffer = false
		}
	}
	return decodedData, totalErrors
}
