# go-gp
[![Go Report Card](https://goreportcard.com/badge/github.com/Ficoto/go-gp)](https://goreportcard.com/report/github.com/Ficoto/go-gp)
[![MIT license](http://img.shields.io/badge/license-MIT-9d1f14)](http://opensource.org/licenses/MIT)

golang实现的协程池，特性是快速增加协程，协程复用，闲置超时回收协程，调用Close()会等待当前正在执行中的协程执行完之后才会回调

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
	p.Run()
	p.Go(func() error {
		// doing something
		return nil
	})
	p.Close()
}
```

## 已知问题
 - 在调用Close方法时，待开启的新任务会直接被抛弃，仅仅只有执行中的任务会保障被执行完
 - Close方法不会等待所有的任务callback执行完，关于要不要等待callback执行完再回调，还在思考中