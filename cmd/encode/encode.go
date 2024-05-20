package encode

import (
	"log"
	"math/rand"
	"time"
)

// Порождающий полином в бинарном представлении
const generator = 0b1011

// Функция для деления нашего полинома на порождающий полином generator
func divide(block byte, generator byte) byte {
	for i := 6; i >= 3; i-- {
		if (block>>i)&1 != 0 {
			block ^= generator << (i - 3)
		}
	}
	return block
}

// Функция для кодирования 4-битного блока
func blockEncode(block byte) byte {
	//log.Println("Block before encode", block)
	// Сдвигаем блок на 3 позиции влево, чтобы освободить место для контрольных битов
	originalBlock := block << 3

	// Выполняем деление и получаем остаток
	remainder := divide(originalBlock, generator)

	// Прибавляем остаток к исходному блоку
	encodedBlock := originalBlock | remainder

	//log.Println("Block after encoded", encodedBlock)

	return encodedBlock
}

// DataEncode Функция для кодирования всех данных
func DataEncode(data []byte) []byte {
	var encodedData []byte
	for _, b := range data {
		// Кодируем первую половину байта (первые 4 бита - 0x0F = 1111)
		block1 := (b >> 4) & 0x0F
		encodedBlock1 := blockEncode(block1)
		encodedData = append(encodedData, encodedBlock1)

		// Кодируем вторую половину байта (вторые 4 бита)
		block2 := b & 0x0F
		encodedBlock2 := blockEncode(block2)
		encodedData = append(encodedData, encodedBlock2)
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Вносим ошибку с вероятностью 10% в один случайный бит из всего массива encodedData
	if rng.Float64() < 1 {
		log.Println("Without error", encodedData)

		// Выбираем случайный байт в массиве encodedData
		bytePosition := rng.Intn(len(encodedData))
		// Выбираем случайный бит в выбранном байте
		bitPosition := rng.Intn(8)
		// Инвертируем выбранный бит
		encodedData[bytePosition] ^= 1 << bitPosition
		log.Println("Error number ", bytePosition, "error ", bitPosition)
		log.Println("Error appeared")
		log.Println("With error", encodedData)
		log.Println()
	}

	return encodedData
}
