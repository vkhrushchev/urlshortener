package util

import "fmt"

func ExampleGetShortURL() {
	out1 := GetShortURL("http://localhost", "abc")
	fmt.Println(out1)

	out2 := GetShortURL("http://localhost/", "abc")
	fmt.Println(out2)

	// Output:
	// http://localhost/abc
	// http://localhost/abc
}
