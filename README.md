# go-get-http-chunked
read and process chunked data from http stream


# description
1. get chancked message by read by chunk
2. send data to chain
3. worker gorutine make processing
4. every some time period we make print result
5. gracefull  SIGTERM, SIGHUP and SIGQUIT exit and print final result

# configuration
example file conf.yaml
	 
    url: http://127.0.0.1/stream
    period: 5 # in seconds
     
# installation

	go build .
    # or
    go build -o chunked main.go