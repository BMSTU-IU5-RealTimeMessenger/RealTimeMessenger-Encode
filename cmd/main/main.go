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

	// Создаем новый генератор случайных чисел с сидом на основе текущего времени
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Проверяем вероятность потери кадра
	if rng.Float64() < 0.015 {
		log.Println("Frame lost")
		c.Status(http.StatusOK)
		return
	}

	log.Println("Got request with data:\n", string(data))
	//log.Println(" {\n    \"test\": \"hello\"\n}\n")
	// Один закодированный кадр
	encodedData := encode.DataEncode(data)
	log.Println("EncodeData:\n", string(encodedData))
	// Один исправленный кадр (сегмент)
	decodedData, numberErrors := decode.DataDecode(encodedData)
	log.Println("DecodeData:\n", string(decodedData))
	log.Println("Errors:\n", numberErrors)
	hadErrors := false
	if numberErrors > 0 {
		hadErrors = true
	}

	if err := s.send(data, hadErrors); err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	//time.Sleep(100 * time.Second)
	c.Status(http.StatusOK)
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
	Segment []byte `json:"segment"`
	Error   bool   `json:"error"`
}

func (s *Server) send(data []byte, hadError bool) error {
	payload := Payload{
		Segment: data,
		Error:   hadError,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "http://"+s.Destination+"/take", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

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
