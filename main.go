package main

import (
	// "bytes"
	"context"
	"fmt"
	"time"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"gopkg.in/yaml.v2"
)

type Data struct {
	Data float64 `json:"data"`
}

const (
	TIME = time.Second * 5
	FILENAME = "conf.yaml"
)


type Config struct {
	Url string `yaml:url`
}

var Counter int32
var Sum float64


func timer(ctx context.Context) {
	ticker := time.Tick(TIME)

	for {
		select {
			case <-ctx.Done():
				fmt.Println("timer Done", ctx.Err())
				return 

			case   <- ticker:
				fmt.Println("***** every 5 sec")
				fmt.Printf("result: %0.8f\n---------------------\n", Sum / float64(Counter))

		}
	}
}


func worker (ctx context.Context, 	in chan float64) {
	for {
		select {
			case <-ctx.Done():			
				fmt.Println("worker Done", ctx.Err())
				return 
			case  num := <- in:
				Sum += num
				Counter ++
				fmt.Printf("get: %0.8f number:%d result: %f\n", num, Counter, Sum / float64(Counter))
		}
	}

}


func onSignal(ctx context.Context, cancel func(), finish chan struct{}) {
	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	var nl struct{}
	select {
		case <- sigChan:
			fmt.Printf("==================\n finish result: %f\n", Sum / float64(Counter))
			cancel()
			finish <- nl
		case <- ctx.Done():
			fmt.Println("signal Done")
	}
}


func downloader(ctx context.Context, cn chan float64, cancel func()) {
	var data Data
	var conf Config


	yaml_buf, err := readFile()
	if err != nil {
		fmt.Println("read file error:", err)
		cancel()
		return
	}
	err = yaml.Unmarshal(yaml_buf, &conf)
	if err != nil {
		cancel()
		fmt.Println("parse config file:", err)
		return
	}


	tr := &http.Transport{
		MaxIdleConns:       	10,
		IdleConnTimeout:    	15 * time.Second,
		DisableCompression: 	true,
	}


	client := &http.Client{
		Transport: tr, 
	}

	res, err := client.Get(conf.Url)
	if err != nil {
		ctx.Done()
		return
	}	
	defer res.Body.Close()

	select {

	  case <- ctx.Done():

		fmt.Println("downloader Done", ctx.Err())
	  	return

	  default: 
	  	
		for {

			buf := make([]byte, 128) 
			n,err := res.Body.Read(buf)
			if err != nil {
				fmt.Println(err)
				cancel()
				break
			}
			buf = buf[:n]
			err = json.Unmarshal(buf, &data)
	 	    if err != nil {
		        fmt.Println("JSON error", err, string(buf))
		        cancel()
		        break
		    }
			cn <- data.Data
			buf = nil
		}

	}
	fmt.Println("downloader finish")
}


func readFile() ([]byte, error) {
	file, err := os.Open(FILENAME)
	if err != nil {
	    	// fmt.Println(err)
	    	return nil,err
	}
    defer func() {
        if err = file.Close(); err != nil {
            fmt.Println(err)
        }
	    }()

	return ioutil.ReadAll(file)
}

func main() {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	cn_data := make (chan float64)
    finishCn := make(chan struct{})


	go onSignal(ctx, cancel, finishCn)
	go timer(ctx)
	go worker(ctx, cn_data)
	go downloader(ctx, cn_data, cancel)

	<- finishCn
	fmt.Println("Finish gracefull")

}