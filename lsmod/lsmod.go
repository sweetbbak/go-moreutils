package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

// func tabby() {
// 	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
// 	fmt.Fprintln(writer, "Module\tSize\tUsed\tUsed By")
// 	fmt.Fprintf(writer, "%-19s %8s %s", name, size, used)
// 	writer.Flush()

// }

func main() {
	file, err := os.Open("/proc/modules")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Println("Module                 Size Used Used By")

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), " ")
		name, size, used, usedBy := s[0], s[1], s[2], s[3]
		final := fmt.Sprintf("%-19s %8s %s", name, size, used)
		if usedBy != "-" {
			usedBy = usedBy[:len(usedBy)-1]
			final += fmt.Sprintf(" %s", usedBy)
		}

		fmt.Println(final)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

}
