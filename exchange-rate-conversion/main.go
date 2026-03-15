package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
)

type ConversionResult struct {
	Currency string
	Amount   float64
}

var rate map[string]float64 = map[string]float64{
	"USD": 15000,
	"EUR": 16000,
	"JPY": 140,
	"SGD": 11000,
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var amountRupiah float64
	result := []ConversionResult{}

	// prepare data
	for {
		fmt.Print("Input the nominal in Rupiah: ")
		scanner.Scan()
		text := scanner.Text()
		n, err := strconv.Atoi(text)
		if err == nil && n > 0 {
			amountRupiah = float64(n)
			break
		}
		fmt.Println("\nInvalid nominal. Enter the valid nominal.")
	}

	// show data state
	fmt.Println(amountRupiah)
	fmt.Printf("Nominal you can get for %.3f IDR:\n", amountRupiah)

	// collect data after conversion
	for currency, denom := range rate {
		result = append(result, ConversionResult{
			Currency: currency,
			Amount:   amountRupiah / denom,
		})
	}

	// show result
	for _, data := range result {
		fmt.Printf("%s : %.3f\n", data.Currency, data.Amount)
	}
}
