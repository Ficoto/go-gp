# go-gp
[![Go Report Card](https://goreportcard.com/badge/github.com/Ficoto/go-gp)](https://goreportcard.com/report/github.com/Ficoto/go-gp)
[![MIT license](http://img.shields.io/badge/license-MIT-9d1f14)](http://opensource.org/licenses/MIT)

golang实现的协程池

## Usage
```
go get github.com/Ficoto/go-gp
```
Create a Pool to use
```go
package main

import "github.com/Ficoto/go-gp"

func main() {
	p := gp.New()
	p.Go(func() {
		// doing something
	})
	p.Close()
}
```

## 特性
 - 允许设置异步任务重试
 - 允许设置异步任务执行完成回调方法
 - 调用close方法会等待已提交的任务执行完且回调完才会结束
 - 并发协程到设置的最大协程数时，提交新任务会被堵塞
 - 捕获任务执行panic并将panic栈信息作为error传递给callback