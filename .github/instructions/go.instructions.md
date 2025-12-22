---
applyTo: '**/*.go'
---
<!-- https://atomgit.com/lingma/lingma-project-rule-template/blob/master/golang/project_rule.md -->

# Go 语言编码规范

## 具体开发规范

### 初始化

slice 和 map 在声明时，推荐使用 `make` 进行初始化，而不是使用字面量初始化。

### 循环

对切片进行循环的时候，避免对大结构体进行循环，而应该使用迭代器模式。

### 错误处理

错误处理应该使用 `errors` 包，而不是使用 `panic`。

### 注释规范

#### 接口和实现的注释

方法功能的注释应该出现在接口定义处，而不是在实现处。如果方法实现的注释出现在实现处，可能会导致注释和接口定义脱节，降低代码的可读性和维护性。

实现的方法注释应该使用 `Implements` 标记，明确指出该方法实现了哪个接口的方法。这有助于提高代码的可读性和可维护性。

```go
type AInterface interface {
    // DoSomething 执行某个操作
    DoSomething(param int) error
}

type AStruct struct {
    // 内部字段
}

// DoSomething Implements [AInterface.DoSomething]
func (s *AStruct) DoSomething(param int) error {
    // 方法实现
    return nil
}
```

## 参考

- [Google Go 语言编码规范](https://google.github.io/styleguide/go/)
- [Go 语言最佳实践](https://google.github.io/styleguide/go/best-practices)
