package main

import (
	"log"
	"yikou-ai-go-microservice/services/screenshot/handler"
	screenshot "yikou-ai-go-microservice/services/screenshot/kitex_gen/screenshotservice"
)

func main() {
	go func() {
		svr := screenshot.NewServer(new(handler.ScreenshotServiceImpl))

		err := svr.Run()

		if err != nil {
			log.Println(err.Error())
		}
	}()
}
