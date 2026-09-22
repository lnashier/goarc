module httpservice

go 1.26.0

replace github.com/lnashier/goarc/v2 => ../../../goarc

require (
	github.com/gorilla/mux v1.8.1
	github.com/lnashier/goarc/v2 v2.0.0
)

require github.com/urfave/negroni v1.0.0 // indirect
