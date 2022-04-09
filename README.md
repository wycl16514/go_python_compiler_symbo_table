本节我们要完成一个任务，给定如下一段代码：
```
{int x; char y; {bool y; x; y;} x; y;}
```
解析上面代码后输出结果为：
```
{{x:int; y:bool} x:int; y:char}
```

大家如果对c, c++, java有所了解，那么就会知道作用域这个概念。所谓作用域就是变量在一个范围内起作用，一旦出了既定范围，那么它就会失效。c,c++,java用{表示作用域的起始，用}表示作用域的结束。内层作用域的变量会覆盖上一层作用域的变量。例如在上面代码中最外层定义了两个变量，分别是int类型的x，和char类型的y,在内层作用域又定义了一个bool类型的同名变量y，它会覆盖外出的char类型y，在内层作用域访问y时，我们访问的是类型为bool的y，但由于内层作用域没有定义x，因此访问x时，它对应外层作用域的x，因此我们的任务是识别作用域，同时解析出变量在不同作用域中对应的类型。

在编译原理中，使用一种叫符号表的特殊结构来记录变量的信息，例如变量的类型，名称，在内存中的地址等。在使用IDE开发代码时，我们调试时，将鼠标挪到某个变量名称上，IDE就会显示出变量的值等信息，这些信息就得依靠符号表来存储，没有符号表就不能实现断点或是单步调试。

在代码解析过程中，一旦发现有变量定义出现时，编译器就跟构造一条符号记录，然后将其插入到符号表。当编译器发现代码进入新的作用域时，它会创建一个新的符号表用于记录新作用域下的变量信息，于是每个作用域都会对应一个符号表，在该作用域下变量的相关信息就从对应符号表查询。内部作用域对应的符号表会有一个指针指向它上一层作用域的符号表，在解析内部作用域的变量时，如果发现某个变量没有出现在其符号表中，那么就顺着指针在上一层符号表查找，如果还是查找不到那么继续往上查找，如果到达最外层作用域，其符号表还是没有对应变量，那么就产生了语法错误，也就是代码使用了未声明的变量，其基本逻辑如下图所示：
![请添加图片描述](https://img-blog.csdnimg.cn/cc0d18b3506d403bb6d02bfd3952f274.png?x-oss-process=image/watermark,type_d3F5LXplbmhlaQ,shadow_50,text_Q1NETiBAdHlsZXJfZG93bmxvYWQ=,size_11,color_FFFFFF,t_70,g_se,x_16)

从上图看到，前面代码中最内层的作用域访问了变量x，但是x并没有在当前作用域里定义，于是编译器从当前作用域对应的符号表指针出发，找到上一层作用域的符号表，在那里查询到了x的定义，因此在内存作用域中使用的x，对应为外层作用域定义的x。理论说的太多容易糊涂，我们看看具体的代码实现，在Parser目录下新增symbol.go,添加如下代码：
```
package parser

type Symbol struct {
	VariableName string 
	Type   string 
}

func NewSymbol (name , var_type string) *Symbol {
	return &Symbol {
		VariableName: name, 
		Type:  var_type,
	}
}
```
这里定义的Symbol对象比较简单，它只记录了当前变量名称和类型，根据前面的任务，我们解析代码后，遇到变量表达式例如: x; y;时，只需要将他们在对应作用域内的变量类型输出即可。下面我们看看上面图中符号表所形成的链表如何构成，添加Env.go文件，输入如下代码：
```
package parser

type Env struct {
	table map[string]Symbol 
	prev  *Env  //这里形成链表
}

func NewEnv(p *Env) *Env {
	return &Env {
		table : new(map[string]Symbol),
		prev : p,
	}
}

func (e *Env)Put(s string, sym Symbol) {
	e.table[s] = sym 
}

func (e *Env)Get(s string) *Symol {
	//查询变量符号时，如果当前符号表没有定义，我们要往上一层作用域做进一步查询
	for e := s; e != nil; e = e.prev {
		found, ok := e.table[s]
		if ok {
			return found 
		}
	}

	return nil 
}
```
Env对应的就是符号表，它使用一个哈希表存储变量对应的符号，也就是Symbol类，当查询变量对应符号时，它现在自己的哈希表中查询，如果查询不到，它通过prev指针找到上一层的符号表，然后继续查询，如果所有作用域的符号表都找不到对应的符号，那么说明代码出错，它引用了一个未定义的变量。接下来我们看看语法解析解析部分，首先我们看看语法定义：
```
prggram ->  block  {top = nil}
block -> '{'  decls stmts '}'  {saved = top; top = NewEnv(top); print("{"}
decls -> decls decl | ε
decl -> type id ";"  {s = NewSymbol(type.lexeme, id.lexeme); top[id.lexeme] = s}
stmts -> stmts stmt | ε
stmt -> block | factor ";"  {print(";")}
factor -> id  {s = top[id.lexeme]; print(id.lexeme); print(";"); print(s.type.lexeme);}
```

我们看看语法的定义，progrom表示整个函数，它分解为block,后者表示一个有一对大括号包括在一起的代码块，top指向当前作用域对应的Env对象，在程序开始解析时先把它设置为nil。在解析block时，首先判断它是否以左大括号"{"开始，然后跟着解析一系列变量声明，类似于"int x;" , "bool y;" 等语句都是变量声明，这些语句对应的就是decl,一系列变量声明语句合在一起就对应decls，当然声明可以是空语句，例如单单一个分号";"也算是变量声明，它是空声明。


变量声明用可以分解成type和id的组合，type 字符串"int" , "float", "bool", "char"等，id对应的就是变量名，也就是identifier，stmts表示的是多个表达式语句，其中单个表达式语句用stmt表示，由于我们这解析的表达式就是一个变量名加一个分号这么简单，于是stmt可以分解成factor加分号，然后factor再分解成一个id，于是stmt其实就是指"x;", "y;"这类语句，同时一个表达式又可以对应一个作用域区块，于是它又能分解成block，这样就能形成嵌套的作用域，也就是一个大括号程序块内部又能有一个大括号程序块。

在上面语法表达式中，有两个表达式出现了左递归，根据前面章节描述的消除方法，他们改为：
```
decls -> decls_r 
decls_r -> decl decls_r | ε

stmts -> stmts stmt
stmts-> stmts_r 
stmts_r -> stmt stmts_r | ε
```

我们还是通过代码来解读上面的语法解析更方便，修改simple_parser.go如下：
```
package simple_parser

import (
	"errors"
	"fmt"
	"lexer"
)

type SimpleParser struct {
	lexer lexer.Lexer
	top   *Env
	saved *Env
}

func NewSimpleParser(lexer lexer.Lexer) *SimpleParser {
	return &SimpleParser{
		lexer: lexer,
		top:   nil,
		saved: nil,
	}
}

func (s *SimpleParser) Parse() error {
	return s.program()
}

func (s *SimpleParser) program() error {
	s.top = nil
	return s.block()
}

func (s *SimpleParser) match(str string) error {
	if s.lexer.Lexeme != str {
		err_s := fmt.Sprintf("match error, expect %s got %s ", str, s.lexer.Lexeme)
		return errors.New(err_s)
	}

	return nil
}

func (s *SimpleParser) block() error {
	s.lexer.Scan()
	err := s.match("{")
	if err != nil {
		return err
	}

	//执行语法定义的操作
	s.saved = s.top
	s.top = NewEnv(s.top)
	fmt.Print("{")

	err = s.decls()
	if err != nil {
		return err
	}

	err = s.stmts()
	if err != nil {
		return err
	}

	err = s.match("}")
	if err != nil {
		return err
	}

	//执行语法定义中的操作
	s.top = s.saved
	fmt.Print("}")
	return nil
}

func (s *SimpleParser) decls() error {
	return s.decls_r()
}

func (s *SimpleParser) decls_r() error {
	var err error
	tag, err := s.lexer.Scan()
	if err != nil {
		return err
	}
	if tag.Tag == lexer.TYPE {
		//遇到int, float等变量定义字符串，因此解析变量定义
		s.lexer.ReverseScan()
		err = s.decl()
		if err != nil {
			return err
		}

		return s.decls_r()
	} else {
		//什么都不做 , 对应空
		s.lexer.ReverseScan()
	}

	return nil
}

func (s *SimpleParser) decl() error {
	tag, err := s.lexer.Scan()
	if err != nil {
		return err
	}
	if tag.Tag != lexer.TYPE {
		str := fmt.Sprintf("in decl, expect type definition but got: %s", s.lexer.Lexeme)
		return errors.New(str)
	}
	type_str := s.lexer.Lexeme

	tag, err = s.lexer.Scan()
	if err != nil {
		return err
	}
	if tag.Tag != lexer.ID {
		str := fmt.Sprintf("in decl, expect identifier, but got: %s", s.lexer.Lexeme)
		return errors.New(str)
	}
	id_str := s.lexer.Lexeme
	//执行语法定义中的操作
	symbol := NewSymbol(id_str, type_str)
	s.top.Put(id_str, symbol)

	_, err = s.lexer.Scan()
	if err != nil {
		return err
	}

	err = s.match(";")

	return err
}

func (s *SimpleParser) stmts() error {
	/*
		消除左递归 stmts -> stmts stmt | epsilon
		stmts -> epsilon r_stmts
		r_stmts -> stmt r_stmts | epsilon
	*/

	return s.r_stmts()
}

func (s *SimpleParser) r_stmts() error {
	tag, err := s.lexer.Scan()
	if err != nil {
		return err
	}
	if tag.Tag == lexer.ID || tag.Tag == lexer.LEFT_BRACE {
		s.lexer.ReverseScan()
		err = s.stmt()
		if err != nil {
			return err
		}

		err = s.r_stmts()
	} else if tag.Tag == lexer.SEMICOLON {
		return nil
	}

	return nil
}

func (s *SimpleParser) stmt() error {
	tag, err := s.lexer.Scan()
	if err != nil {
		return err
	}
	if tag.Tag == lexer.LEFT_BRACE {
		s.lexer.ReverseScan()
		err = s.block()
	} else if tag.Tag == lexer.ID {
		s.lexer.ReverseScan()
		err = s.factor()
		s.lexer.Scan()
		err = s.match(";")
		//执行语法定义的操作
		if err == nil {
			fmt.Print("; ")
		}
	} else {
		err = errors.New("stmt parsing error")
	}

	return err
}

func (s *SimpleParser) factor() error {
	tag, err := s.lexer.Scan()
	if err != nil {
		return err
	}
	if tag.Tag != lexer.ID {
		str := fmt.Sprintf("expect identifier , got %s ", s.lexer.Lexeme)
		return errors.New(str)
	}

	//执行语法定义的操作
	symbol := s.top.Get(s.lexer.Lexeme)

	fmt.Print(s.lexer.Lexeme)
	fmt.Print(":")
	fmt.Print(symbol.Type)

	return nil
}

```
从上面代码中可以看到，语法解析的函数调用顺序基本上依照语法表达式的描述，同时需要注意的是，代码在执行完解析后会执行语法表达式中的”行为“，也就是执行对应的代码逻辑，最后我们修改main.go:
```
package main

import (
	"fmt"
	"io"
	"lexer"
	"simple_parser"
)

func main() {
	source := "{int x; char y; {bool y; x; y;} x; y;}"
	my_lexer := lexer.NewLexer(source)
	parser := simple_parser.NewSimpleParser(my_lexer)
	err := parser.Parse()
	if err == io.EOF || err == nil {
		fmt.Println("parsing success")
	} else {
		fmt.Println("source is ilegal : ", err)
	}
}

```
完成代码后，所得结果如下：
![请添加图片描述](https://img-blog.csdnimg.cn/2fddc07431304100a4e73be44338185c.png?x-oss-process=image/watermark,type_d3F5LXplbmhlaQ,shadow_50,text_Q1NETiBAdHlsZXJfZG93bmxvYWQ=,size_20,color_FFFFFF,t_70,g_se,x_16)
从上图看，输出结果与开头预期是一致的
