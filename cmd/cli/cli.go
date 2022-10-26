package main

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"gophkeeper/cmd/cli/client"
	"gophkeeper/cmd/cli/ui"
	"gophkeeper/internal/domain"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	cli         = kingpin.New("Gophkeeper", "Gophkeeper is a tool for keeping credentials, text data, credit cards and binary data.")
	cmd         string
	accessToken = cli.Flag("access-token", "access token is needed to call private methods").String()
	serverAddr  = cli.Flag("server-addr", "server address").Default("http://localhost:8081").String()

	interactiveMode = cli.Command("interact", "run client in interactive mode")

	//auth
	signUp         = cli.Command("sign-up", "create new user and login in service")
	signUpLogin    = signUp.Flag("login", "login or user name").Required().String()
	signUpPassword = signUp.Flag("password", "password for user name").Required().String()

	signIn         = cli.Command("sign-in", "enter the service")
	signInLogin    = signIn.Flag("login", "login or user name").Required().String()
	signInPassword = signIn.Flag("password", "password for user name").Required().String()

	refresh      = cli.Command("refresh", "refresh access and refresh tokens")
	refreshToken = refresh.Flag("refresh-token", "refresh tokens").Required().String()

	//text
	textGetAll = cli.Command("text-get", "get all text data from server")

	textUpdateByID = cli.Command("text-update", "update record")
	textUID        = textUpdateByID.Flag("id", "record id").Required().Int()
	textUText      = textUpdateByID.Flag("text", "text").Required().String()
	textUMetadata  = textUpdateByID.Flag("metadata", "metadata").Required().String()

	textCreateNew = cli.Command("text-create", "crate new record")
	textCText     = textCreateNew.Flag("text", "text").Required().String()
	textCMetadata = textCreateNew.Flag("metadata", "metadata").Required().String()

	//card
	cardGetAll = cli.Command("card-get", "get all card data from server")

	cardUpdateByID  = cli.Command("card-update", "update record")
	cardUID         = cardUpdateByID.Flag("id", "record ID").Required().Int()
	cardUCardNumber = cardUpdateByID.Flag("card-number", "16-digit card number").Required().String()
	cardUExpDate    = cardUpdateByID.Flag("exp-date", "expired date in format mm/yy").Required().String()
	cardUCVV        = cardUpdateByID.Flag("cvv", "cvv").Required().String()
	cardUName       = cardUpdateByID.Flag("name", "name").Required().String()
	cardUSurname    = cardUpdateByID.Flag("surname", "surname").Required().String()
	cardUMetadata   = cardUpdateByID.Flag("metadata", "metadata").Required().String()

	cardCreateNew   = cli.Command("card-create", "crate new record")
	cardCCardNumber = cardCreateNew.Flag("card-number", "16-digit card number").Required().String()
	cardCExpDate    = cardCreateNew.Flag("exp-date", "expired date in format mm/yy").Required().String()
	cardCCVV        = cardCreateNew.Flag("cvv", "cvv").Required().String()
	cardCName       = cardCreateNew.Flag("name", "name").Required().String()
	cardCSurname    = cardCreateNew.Flag("surname", "surname").Required().String()
	cardCMetadata   = cardCreateNew.Flag("metadata", "metadata").Required().String()

	//cred
	credGetAll = cli.Command("cred-get", "get all cred data from server")

	credUpdateByID = cli.Command("cred-update", "update record")
	credUID        = credUpdateByID.Flag("id", "record id").Required().Int()
	credULogin     = credUpdateByID.Flag("login", "login").Required().String()
	credUPassword  = credUpdateByID.Flag("password", "password").Required().String()
	credUMetadata  = credUpdateByID.Flag("metadata", "metadata").Required().String()

	credCreateNew = cli.Command("cred-create", "crate new record")
	credCLogin    = credCreateNew.Flag("login", "login").Required().String()
	credCPassword = credCreateNew.Flag("password", "password").Required().String()
	credCMetadata = credCreateNew.Flag("metadata", "metadata").Required().String()
)

func init() {
	for i, arg := range os.Args {
		if strings.HasPrefix(arg, "--") {
			split := strings.SplitN(arg, "=", 2)
			split[0] = strings.ReplaceAll(split[0], "_", "-")
			os.Args[i] = strings.Join(split, "=")
		}
	}

	cmd = kingpin.MustParse(cli.Parse(os.Args[1:]))
}

func main() {
	if cmd == interactiveMode.FullCommand() {
		err := tea.NewProgram(ui.NewMainModel(*serverAddr)).Start()
		if err != nil {
			fmt.Println(err.Error())
		}
	} else {
		run()
	}
}

func run() {
	ctx, cancel := context.WithCancel(context.Background())
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-signalChan
		cancel()
	}()

	cl := client.NewGKClient(*serverAddr)

	if *accessToken == "" { //open methods
		switch cmd {
		case signUp.FullCommand():
			tokens, err := cl.UserSignUp(ctx, client.AuthInput{
				Login:    *signUpLogin,
				Password: *signUpPassword,
			})
			if err != nil {
				fmt.Println(err.Error())
			}
			fmt.Printf("Access Token: %s\nRefresh token: %s\n", tokens.AccessToken, tokens.RefreshToken)
		case signIn.FullCommand():
			if tokens, err := cl.UserSignIn(ctx, client.AuthInput{
				Login:    *signInLogin,
				Password: *signInPassword,
			}); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Printf("Access Token: %s\nRefresh token: %s\n", tokens.AccessToken, tokens.RefreshToken)
			}
		case refresh.FullCommand():
			if tokens, err := cl.UserRefresh(ctx, *refreshToken); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Printf("Access Token: %s\nRefresh token: %s\n", tokens.AccessToken, tokens.RefreshToken)
			}
		default:
			fmt.Println("no such open methods")
		}
	} else { //closed methods
		cl.SetAccessToken(*accessToken)
		switch cmd {
		case textGetAll.FullCommand():
			if texts, err := cl.GetAllTextData(ctx); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Printf("%-3s %-23.23s %-16s\n", "ID", "Text", "Metadata")
				for _, text := range texts {
					fmt.Printf("%-3d %-23.23s %-16s", text.ID, text.Text, text.Metadata)
				}
			}
		case textUpdateByID.FullCommand():
			if err := cl.UpdateTextData(ctx, domain.TextData{
				ID:       *textUID,
				Text:     *textUText,
				Metadata: *textUMetadata,
			}); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Text data was successfully updated")
			}
		case textCreateNew.FullCommand():
			if err := cl.CreateNewTextData(ctx, domain.TextData{
				Text:     *textCText,
				Metadata: *textCMetadata,
			}); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Text data was successfully added")
			}
		//********************************************************************
		case cardGetAll.FullCommand():
			if cards, err := cl.GetAllCardData(ctx); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Printf("%-3s %-23.23s %-6s %-4s %-15s %-15s  %-15s\n", "ID", "CardNumber", "ExpDate", "CVV", "Name", "Surname", "Metadata")
				for _, card := range cards {
					fmt.Printf("%-3d %-33.33s %-8s %-4s %-15s %-15s %-15s",
						card.ID,
						card.CardNumber,
						card.ExpDate,
						card.CVV,
						card.Name,
						card.Surname,
						card.Metadata,
					)
				}
			}
		case cardUpdateByID.FullCommand():
			if err := cl.UpdateCardData(ctx, domain.CardData{
				CardNumber: *cardUCardNumber,
				ExpDate:    parseExpireDate(*cardUExpDate),
				CVV:        *cardUCVV,
				Name:       *cardUName,
				Surname:    *cardUSurname,
				Metadata:   *cardUMetadata,
			}); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Card data was successfully updated")
			}
		case cardCreateNew.FullCommand():
			if err := cl.UpdateCardData(ctx, domain.CardData{
				CardNumber: *cardCCardNumber,
				ExpDate:    parseExpireDate(*cardCExpDate),
				CVV:        *cardCCVV,
				Name:       *cardCName,
				Surname:    *cardCSurname,
				Metadata:   *cardCMetadata,
			}); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Card data was successfully added")
			}
			//********************************************************************
		case credGetAll.FullCommand():
			if creds, err := cl.GetAllCredsData(ctx); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Printf("%-3s %-20s %-20s %-20s\n", "ID", "Login", "Password", "Metadata")
				for _, cred := range creds {
					fmt.Printf("%-3d %-20s %-20s %-20s\n", cred.ID, cred.Login, cred.Password, cred.Metadata)
				}
			}
		case credUpdateByID.FullCommand():
			if err := cl.UpdateCredData(ctx, domain.CredData{
				ID:       *credUID,
				Login:    *credULogin,
				Password: *credUPassword,
				Metadata: *credUMetadata,
			}); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Cred data was successfully updated")
			}
		case credCreateNew.FullCommand():
			if err := cl.CreateNewCredData(ctx, domain.CredData{
				Login:    *credCLogin,
				Password: *credCPassword,
				Metadata: *credCMetadata,
			}); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Cred data was successfully added")
			}
		}
	}
}

func parseExpireDate(exp string) time.Time {
	date, err := time.Parse("01/06", exp)
	if err != nil {
		log.Panicln(err.Error())
	}

	return date
}
