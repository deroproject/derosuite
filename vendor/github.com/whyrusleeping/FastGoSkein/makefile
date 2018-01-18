all:
	go build -gcflags -B -o main

opti:
	gccgo main.go skein.go skein_consts.go -O3 -o main
