package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	fmt.Println(os.Args)

	management := flag.String("management", "", "")
	remoteAddr := flag.String("remote", "", "")
	port := flag.Int("port", 0, "")
	queryPasswords := flag.Bool("management-query-passwords", false, "")
	flag.Parse()

	fmt.Println("These are values you passed to me: ", *management, *remoteAddr, *port, *queryPasswords)

	fmt.Println("And I am done here - bye")
}
