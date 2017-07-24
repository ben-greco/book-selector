package main

import (
	"bufio"
	"fmt"
	"github.com/spf13/viper"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	reader    = bufio.NewReader(os.Stdin)
	text      string
	bookMap   = make(map[int]string)
	voterSet  = make(map[string]bool)
	bookSlice []string
	display   = true
)

func main() {

	// Clear the screen.
	for i := 0; i < 100; i++ {
		fmt.Println()
	}

	// Present a welcome message.
	fmt.Print("\nWelcome to book selector!\n")

	// Complete this loop until exit is indicated.
	for text != "X" {

		// Print loop summary and instructions.
		if display {
			printBookVotes()
			fmt.Print("\nSelect one of the following options: \n\na/A: Add the selections of a voter. " +
				"\nr/R: Read in the list of books. \nselect: Select a book randomly based on the " +
				"given votes and finish program (select must be typed completely) \nx: Exit Book " +
				"Selector.\n\nWhat would you like do next? \n")
		}

		// Receive selection and clean it.
		text, _ = reader.ReadString('\n')
		text = strings.ToUpper(text)
		text = strings.Trim(text, " \n")
		fmt.Print(strings.Repeat("\n", 100))
		switch text {
		case "A":
			addVotes()
			display = true
		case "R":
			readInBooks()
			display = true
		case "SELECT":
			display = false
			selectAndDisplayBook()
		case "X":
			fmt.Print("\n\nNow exiting Book Selector!\n\n")
		default:
			display = true
			fmt.Print("That isn't an option that Book Selector understands. Please try again!\n")
		}

	}
}

func selectAndDisplayBook() {

	line := strings.Repeat("~", 15)

	fmt.Println("\n" + strings.Repeat(line+"\n", 2))

	for i, book := range bookSlice {
		fmt.Printf("%d.) %s\n", i, book)
	}

	fmt.Println("\n" + strings.Repeat(line+"\n", 2))

	fmt.Print("\n\nSELECION TIME!\n\n")

	fmt.Print("\n" + strings.Repeat(line+"\n", 2))

	printBookVotes()

	fmt.Print("\n" + strings.Repeat(line+"\n", 2))

	fmt.Println()

	magicNumber := rand.Intn(len(bookSlice))

	fmt.Printf("THE MAGIC NUMBER IS: %d\n\n", magicNumber)
	fmt.Printf("WE WILL BE READING: \n\n%s\n\n", bookSlice[magicNumber])

}

func addVotes() {

	// Ask for the name of the next voter and receive and format it.
	fmt.Println("\nWhat is the name of the voter?")
	name, _ := reader.ReadString('\n')
	name = strings.ToUpper(strings.Trim(name, " .!\n"))

	// If this voter has already voted, reject the votes.
	if voterSet[name] {
		fmt.Println("THIS VOTER ALREADY EXISTS! NO ADDITIONS WILL BE MADE")
		return
	}

	// Ask for, receive, and format the votes of this voter.
	fmt.Printf("\nPlease enter %d votes seperated by commas. Your first three votes will receive a weighting of "+
		"%d, %d, and %d respectively.\n\n", viper.GetInt("numvotes"), viper.GetInt("firstvoteweight"),
		viper.GetInt("secondvoteweight"), viper.GetInt("thirdvoteweight"))
	t, _ := reader.ReadString('\n')
	splits := strings.Split(t, ",")
	var numbers []int
	for _, s := range splits {
		i, err := strconv.ParseInt(strings.Trim(s, " !\n"), 10, 32)
		if err != nil {
			fmt.Printf("Error parsing int out of this line %s | %s", s, err)
			return
		}
		numbers = append(numbers, int(i))
	}

	// Check all the votes of the voter.
	if len(numbers) > viper.GetInt("numvotes") {
		fmt.Printf("Error: There were %d votes when we were expecting %d\n", len(numbers),
			viper.GetInt("numvotes"))
	}
	for _, n := range numbers {
		if n < 1 {
			fmt.Printf("\nError: This number is less than or equal to zero: %d", n)
			return
		}
		if n > len(bookMap) {
			fmt.Printf("\nError: This number is greater than the read in number of books: %d\n", n)
			if n == 1 {
				fmt.Print("\n1 IS BIGGER THAN THE NUMBER OF BOOKS SO YOU PROBABLY FORGOT TO READ THE " +
					"BOOK LIST INTO THE PROGRAM\n")
			}
			return
		}
	}

	// Once everything is cleared, actually add all the votes and add the voter to the voter set.
	voteNumber := 0
	var voteWeight int
	for _, n := range numbers {

		// Assign the vote weight based on the vote number.
		switch voteNumber {
		case 0:
			voteWeight = viper.GetInt("firstvoteweight")
		case 1:
			voteWeight = viper.GetInt("secondvoteweight")
		case 2:
			voteWeight = viper.GetInt("thirdvoteweight")
		default:
			fmt.Println("ERROR: This is more votes than we've ever seen before or were expecting!")
		}

		// Add this to bookSlice voteWeight times.
		for i := 0; i < voteWeight; i++ {
			bookSlice = append(bookSlice, bookMap[n])
		}

		// Increment vote number.
		voteNumber++
	}
	voterSet[name] = true

}

func readInBooks() {

	// Open book list from configured location
	f, err := os.Open(viper.GetString("booklistlocation"))
	if err != nil {
		fmt.Printf("Error reading in books from this file: %s | %s", viper.GetString("booklistlocation"), err)
	}
	scanner := bufio.NewScanner(f)

	// Go through each line
	for scanner.Scan() {

		// Read the line and parse it based on the colon delimiter. Present an error if necessary.
		line := scanner.Text()
		splits := strings.Split(line, ":")
		if len(splits) > 2 {
			fmt.Printf("ERROR: this line in the book list contains more than one colon: %s", line)
			return
		}
		if len(splits) == 1 {
			fmt.Printf("ERROR: There was no colon in this line of the file: %s", line)
		}

		// Parse, format, and save the book number and book name in the book map.
		bookNum, err := strconv.ParseInt(strings.Trim(splits[0], " \n"), 10, 32)
		if err != nil {
			fmt.Printf("Error parsing int from this line %s : %s\n", line, err)
		}
		bookName := strings.ToUpper(strings.Trim(splits[1], " \n"))
		bookMap[int(bookNum)] = bookName
	}

	fmt.Println("\nBOOKS HAVE BEEN READ IN")

}

func printBookVotes() {

	// If bookSlice has not been populated with votes for books, do nothing.
	if bookSlice == nil {
		return
	}

	// Create voteMap and fill it with the counts of the number of book votes so far.
	voteMap := make(map[string]int)
	for _, book := range bookSlice {
		voteMap[book]++
	}

	// Print overall message.
	fmt.Printf("\nSo far %d people have voted, they are:\n", len(voterSet))
	for v := range voterSet {
		fmt.Println(v)
	}
	fmt.Print("\nVOTE TOTALS AND PERCENTAGES:\n")
	var ones, twos, threes []string
	for k := range voteMap {
		ones = append(ones, fmt.Sprintf("%s", k))
		twos = append(twos, fmt.Sprintf(" %d out of %d votes,", voteMap[k], len(bookSlice)))
		myFloat := 100.00 * float64(voteMap[k]) / float64(len(bookSlice))
		threes = append(threes, fmt.Sprint(" "+strconv.FormatFloat(myFloat, 'g', 3, 64)+" chance of being "+
			"selected"))

	}
	maxOnes := 0
	maxTwos := 0
	maxThrees := 0
	for i, thing := range ones {
		if len(thing) > maxOnes {
			maxOnes = len(thing)
		}
		if len(twos[i]) > maxTwos {
			maxTwos = len(twos[i])
		}
		if len(threes[i]) > maxThrees {
			maxThrees = len(threes[i])
		}
	}
	for i := range ones {
		fmt.Printf("%s%s:%s%s%s%s\n", ones[i], strings.Repeat(" ", maxOnes-len(ones[i])),
			strings.Repeat(" ", maxTwos-len(twos[i])), twos[i], strings.Repeat(" ", maxThrees-len(threes[i])),
			threes[i])
	}

}

func init() {

	// Create and parse viper configurations.
	viper.AddConfigPath("conf/")
	viper.AddConfigPath(".")
	viper.SetConfigName("book-selector")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Sprintf("Problem reading in config file: %s", err))
	}

	// Seed random number for selection.
	rand.Seed(time.Now().UnixNano())
}
