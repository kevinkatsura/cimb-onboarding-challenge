package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type CardSpec struct {
	prefix []string
	length []int
}

var cardSpecs = map[string]CardSpec{
	"China UnionPay": {
		prefix: []string{"62"},
		length: []int{16, 17, 18, 19},
	},
	"Switch": {
		prefix: []string{"4903", "4905", "4911", "4936", "564182", "633110", "6333", "6759"},
		length: []int{16, 18, 19},
	},
}

func detectCard(cardNumber string) (string, *CardSpec) {
	for name, spec := range cardSpecs {
		for _, prefix := range spec.prefix {
			if strings.HasPrefix(cardNumber, prefix) {
				return name, &spec
			}
		}
	}
	return "", nil
}

func validLength(targetLength int, spec *CardSpec) bool {
	for _, cardLength := range spec.length {
		if cardLength == targetLength {
			return true
		}
	}
	return false
}

func main() {
	var cardNumbers []string
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Input the number:")

	for scanner.Scan() {
		text := scanner.Text()
		text = strings.TrimSpace(text)
		if text == "" {
			break
		}
		cardNumbers = append(cardNumbers, text)
	}

	fmt.Println("The following is the evaluation results")
	for _, cardNumber := range cardNumbers {
		cardNumberLength := len(cardNumber)
		cardName, cardSpec := detectCard(cardNumber)
		if cardSpec == nil || !validLength(cardNumberLength, cardSpec) {
			fmt.Printf("Card type of %s is unrecognized\n", cardNumber)
			continue
		}
		fmt.Printf("Card number %s is %s\n", cardNumber, cardName)
	}
}
