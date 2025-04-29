package main

// Inter ...
type Inter interface {
	LongLineMethod(aaaaaaaaa string, bbbbbbbbb string, cccccccccc string, dddddddddd string, eeeeeeeeee string, ffffffffff string, gggggggggg string) (string, error)
}

func LongLineMethod2(aaaaaaaaa string, bbbbbbbbb string, cccccccccc string, dddddddddd string, eeeeeeeeee string, ffffffffff string, gggggggggg string) (string, error) {
	return "", nil
}

func main() {
	result, error := LongLineMethod2("aaaaaaaaaa", "bbbbbbbbbb", "ccccccccc", "dddddddddd", "eeeeeeeeee", "ffffffffff", "gggggggggg")
	if error != nil {
		panic(error)
	}
	println(result)

}
