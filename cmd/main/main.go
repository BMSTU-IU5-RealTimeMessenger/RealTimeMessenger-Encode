package main

import (
	"bytes"
	"channelLevelProject/cmd/decode"
	"channelLevelProject/cmd/encode"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type Server struct {
	HTTPClient  *http.Client
	Destination string
}

func New() (*Server, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	return &Server{
		HTTPClient:  http.DefaultClient,
		Destination: os.Getenv("TRANSPORT_LAYER_ADDR"),
	}, nil
}

func (s *Server) Run() {
	r := gin.Default()

	r.POST("/code", s.Code)

	r.Run(os.Getenv("IP") + ":" + os.Getenv("CODE_SERVER_PORT"))
	log.Println("Server is running")
}

func (s *Server) Code(c *gin.Context) {
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	c.Status(http.StatusOK)
	log.Println("Got request with data:\n", string(data))

	// Создаем новый генератор случайных чисел с сидом на основе текущего времени
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Проверяем вероятность потери кадра
	if rng.Float64() < 0.015 {
		log.Println("Frame lost")
		return
	}

	//log.Println(" {\n    \"test\": \"hello\"\n}\n")
	// Один закодированный кадр
	encodedData := encode.DataEncode(data)
	//log.Println("EncodeData:\n", encodedData)
	// Один исправленный кадр (сегмент)
	decodedData, numberErrors := decode.DataDecode(encodedData)
	log.Println("DecodeData:\n", string(decodedData))
	//log.Println("Errors:\n", numberErrors)
	hadErrors := false
	if numberErrors > 0 {
		hadErrors = true
	}

	if err := s.send(decodedData, hadErrors); err != nil {
		log.Println("Sending error", err)
		return
	}

	//time.Sleep(100 * time.Second)
}

func main() {
	server, err := New()
	if err != nil {
		log.Fatalln(err)
		return
	}

	server.Run()
}

type Payload struct {
	Segment string `json:"segment"`
	Error   bool   `json:"error"`
}

//type Segment struct {
//	Data   string `json:"data"`
//	Time   int64  `json:"time"`
//	Number int    `json:"number"`
//	Count  int    `json:"count"`
//}

func (s *Server) send(data []byte, hadError bool) error {
	//var segment Segment
	//err := json.Unmarshal(data, &segment)
	//if err != nil {
	//	log.Println(err)
	//	return err
	//}
	strData := string(data)
	resultString := strings.ReplaceAll(strData, "\\ufffd", "")
	resultString = strings.ReplaceAll(strData, "\\ufffd", "")
	//resultString = strings.ReplaceAll(strData, "\ufffdоа", "")
	payload := Payload{
		Segment: resultString,
		Error:   hadError,
	}
	//log.Println("DATA: ", string(data))
	//log.Println(payload.Segment)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "http://"+s.Destination, bytes.NewBuffer(jsonData))
	//log.Println(payload)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	log.Println("json:", string(jsonData))

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("Unexpected response status: " + resp.Status)
	}
	return nil
}
