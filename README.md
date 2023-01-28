# go-gp
[![Go Report Card](https://goreportcard.com/badge/github.com/Ficoto/go-gp)](https://goreportcard.com/report/github.com/Ficoto/go-gp)
[![MIT license](http://img.shields.io/badge/license-MIT-9d1f14)](http://opensource.org/licenses/MIT)

golang实现的协程池，特性是快速增加协程，协程复用，闲置超时自动结束等待的协程

## Usage
```
go get github.com/Ficoto/go-gp
```
Create a Pool to use
```go
p := New()
p.Run()
p.Go(func() error {
// doing something
    return nil
})
```