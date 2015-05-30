rm -rf $1

go run ./cmd/gsmake/gsmake.go -v -root $1 bootstrap:setup $1
