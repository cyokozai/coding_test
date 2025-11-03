package main

import (
	"fmt"
	"github.com/huandu/xstrings"
	"github.com/shopspring/decimal"
)

func main() {
	fmt.Println("Hello, World!")
	fmt.Println(xstrings.Reverse("Hello, World!"))
	d1 := decimal.NewFromFloat(1.1)
	d2 := decimal.NewFromFloat(2.2)
	sum := d1.Add(d2)
	fmt.Println("Sum:", sum)
}
