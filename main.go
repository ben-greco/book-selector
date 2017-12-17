package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.TODO(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("sheets.googleapis.com-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func readGoogleSheetsBooklist() []string {
	ctx := context.Background()

	b, err := ioutil.ReadFile("conf/client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/sheets.googleapis.com-go-quickstart.json
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets Client %v", err)
	}

	// Prints the names and majors of students in a sample spreadsheet:
	// https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit
	spreadsheetID := "1mFAiopHeUfjyeMjVfDFxKGqSh19jBsTrSN5hEO6ppGs"
	readRange := "Suggestions!A2:E"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
	}

	if len(resp.Values) > 0 {
		return formatGoogleSheetsBookList(resp.Values)
	}
	fmt.Print("No data found.")
	return []string{}
}

func formatGoogleSheetsBookList(i [][]interface{}) []string {
	var gsBookList []string
	for _, row := range i {
		// Print columns A and E, which correspond to indices 0 and 4.
		bookStr := fmt.Sprintf("%s", row[1])
		gsBookList = append(gsBookList, fmt.Sprintf("%s: %s", row[0], strings.Replace(bookStr,
			":", "-", -1)))
	}
	return gsBookList
}

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
				"\nr/R: Read in the list of books. \nw/W: Update book list from Google Sheets. \nselect: Select a book randomly based on the " +
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
		case "W":
			writeBooks()
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

func writeBooks() {
	file, err := os.Create(viper.GetString("booklistlocation"))
	if err != nil {
		fmt.Printf("Error writing books from this file: %s | %s", viper.GetString("booklistlocation"), err)
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	lines := readGoogleSheetsBooklist()
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}

	if err = w.Flush(); err != nil {
		panic(err)
	}
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
		ones = append(ones, k)
		twos = append(twos, fmt.Sprintf(" %d out of %d votes,", voteMap[k], len(bookSlice)))
		myFloat := 100.00 * float64(voteMap[k]) / float64(len(bookSlice))
		threes = append(threes, " "+strconv.FormatFloat(myFloat, 'g', 3, 64)+
			" chance of being selected")

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
