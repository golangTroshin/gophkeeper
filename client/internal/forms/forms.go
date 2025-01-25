// Package forms provides the terminal-based user interface (TUI) forms
// for the GophKeeper client application. It handles user authentication,
// master seed setup, data storage, and retrieval using gRPC.
package forms

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golangTroshin/gophkeeper/client/internal/handlers"
	pb "github.com/golangTroshin/gophkeeper/grpc"
	"github.com/manifoldco/promptui"
	"github.com/rivo/tview"
)

var lastForm *tview.Form

// ShowVersionInfo displays the version and build date in a TUI modal
func ShowVersionInfo(app *tview.Application, client pb.GophKeeperServiceClient, version, buildDate string) {
	versionText := fmt.Sprintf("GophKeeper CLI\n\nVersion: %s\nBuild Date: %s", version, buildDate)

	modal := tview.NewModal().
		SetText(versionText).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			authentication(app, client) // Proceed to login/signup after closing
		})

	modal.SetBorder(true).SetTitle("Version Info").SetTitleAlign(tview.AlignCenter)
	app.SetRoot(modal, true).SetFocus(modal)
}

// Authentication displays the login/signup form to authenticate the user.
func authentication(app *tview.Application, client pb.GophKeeperServiceClient) {
	form := tview.NewForm()
	form.AddInputField("Username", "", 20, nil, nil)
	form.AddPasswordField("Password", "", 20, '*', nil)
	form.AddButton("Login", func() {
		username := form.GetFormItemByLabel("Username").(*tview.InputField).GetText()
		password := form.GetFormItemByLabel("Password").(*tview.InputField).GetText()

		if username == "" {
			errorModal(app, "Username cannot be empty")
			return
		}
		if password == "" {
			errorModal(app, "Password cannot be empty")
			return
		}

		if err := handlers.Login(client, username, password); err != nil {
			errorModal(app, err.Error())
			return
		}
		actionTypeSelection(app, client)
	})
	form.AddButton("Sign Up", func() {
		username := form.GetFormItemByLabel("Username").(*tview.InputField).GetText()
		password := form.GetFormItemByLabel("Password").(*tview.InputField).GetText()

		if username == "" {
			errorModal(app, "Username cannot be empty")
			return
		}
		if password == "" {
			errorModal(app, "Password cannot be empty")
			return
		}

		setupMasterSeed(app, client, username, password)
	})
	form.AddButton("Quit", func() {
		app.Stop()
	})

	form.SetBorder(true).SetTitle("Login/Sign Up").SetTitleAlign(tview.AlignLeft)
	app.SetRoot(form, true)
	lastForm = form
}

// setupMasterSeed prompts the user to enter a master seed for encrypting stored data.
func setupMasterSeed(app *tview.Application, client pb.GophKeeperServiceClient, username, password string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Check if user already exists
	res, err := client.UserExists(ctx, &pb.UserExistsRequest{Username: username})
	if err != nil {
		errorModal(app, fmt.Sprintf("Failed to check user existence: %v", err))
		return
	}

	if res.Exists {
		errorModal(app, "User already exists! Try logging in instead.")
		return
	}

	// Proceed with Master Seed setup if user does not exist
	form := tview.NewForm()
	form.AddInputField("Master Seed", "", 32, nil, nil)

	form.AddButton("Save", func() {
		seed := form.GetFormItemByLabel("Master Seed").(*tview.InputField).GetText()
		if seed == "" {
			errorModal(app, "Master seed cannot be empty")
			return
		}

		if err := handlers.SignUp(client, username, password, seed); err != nil {
			errorModal(app, err.Error())
			return
		}

		actionTypeSelection(app, client)
	})

	form.AddButton("Back", func() {
		authentication(app, client)
	})

	form.SetBorder(true).SetTitle("Set Up Master Seed").SetTitleAlign(tview.AlignLeft)
	app.SetRoot(form, true)
}

// actions defines possible user operations (Save or Retrieve data).
var actions = map[string]uint{
	"save": 1,
	"get":  2,
}

// actionTypeSelection allows the user to choose between saving or retrieving data.
func actionTypeSelection(app *tview.Application, client pb.GophKeeperServiceClient) {
	form := tview.NewForm()
	form.AddButton("Save new data", func() { dataTypeSelection(app, client, actions["save"]) })
	form.AddButton("Get your data", func() { dataTypeSelection(app, client, actions["get"]) })
	form.AddButton("Logout", func() { authentication(app, client) })

	form.SetBorder(true).SetTitle("What do you want to do?").SetTitleAlign(tview.AlignLeft)
	app.SetRoot(form, true)
	lastForm = form
}

// dataTypeSelection allows the user to choose the type of data to store or retrieve.
func dataTypeSelection(app *tview.Application, client pb.GophKeeperServiceClient, actionType uint) {
	form := tview.NewForm()
	form.AddButton("Login/Password Pair", func() { handleDataAction(app, client, pb.DataType_CREDENTIALS, actionType) })
	form.AddButton("Text Data", func() { handleDataAction(app, client, pb.DataType_TEXT, actionType) })
	form.AddButton("Binary Data", func() { handleDataAction(app, client, pb.DataType_BINARY, actionType) })
	form.AddButton("Card Data", func() { handleDataAction(app, client, pb.DataType_CARD, actionType) })
	form.AddButton("Back", func() { actionTypeSelection(app, client) })
	form.AddButton("Logout", func() { authentication(app, client) })

	form.SetBorder(true).SetTitle("Select Data Type").SetTitleAlign(tview.AlignLeft)
	app.SetRoot(form, true)
	lastForm = form
}

// handleDataAction routes the user to save or retrieve data based on actionType.
func handleDataAction(app *tview.Application, client pb.GophKeeperServiceClient, dataType pb.DataType, actionType uint) {
	switch actionType {
	case actions["save"]:
		saveData(app, client, dataType, actionType)
	case actions["get"]:
		getData(app, client, dataType, actionType)
	}
}

// saveData provides a form to save data of the specified type.
func saveData(app *tview.Application, client pb.GophKeeperServiceClient, dataType pb.DataType, actionType uint) {
	form := tview.NewForm()

	switch dataType {
	case pb.DataType_CREDENTIALS:
		form.AddInputField("Login", "", 20, nil, nil)
		form.AddPasswordField("Password", "", 20, '*', nil)
	case pb.DataType_TEXT:
		form.AddInputField("Text", "", 100, nil, nil)
	case pb.DataType_BINARY:
		selectedFilePath, err := promptForFilePath()
		if err != nil {
			errorModal(app, fmt.Sprintf("File selection error: %v", err))
			return
		}
		form.AddInputField("Selected File", selectedFilePath, 100, nil, nil)
	case pb.DataType_CARD:
		form.AddInputField("Card Number", "", 20, nil, nil)
		form.AddInputField("Expiration Date", "", 10, nil, nil)
		form.AddInputField("CVV", "", 3, nil, nil)
	}

	form.AddInputField("Description", "", 100, nil, nil)

	form.AddButton("Save", func() {
		data := handlers.CollectFormData(form, dataType)

		if dataType == pb.DataType_BINARY {
			fileBytes, err := readBinaryFile(data["file_path"])
			if err != nil {
				errorModal(app, fmt.Sprintf("Failed to read file: %v", err))
				return
			}
			data["file_data"] = string(fileBytes)
		}

		err := handlers.SaveData(client, app, dataType, data)
		if err != nil {
			errorModal(app, fmt.Sprintf("Failed to save data: %v", err))
			return
		}
		dataTypeSelection(app, client, actionType)
	})

	form.AddButton("Back", func() { dataTypeSelection(app, client, actionType) })

	form.SetBorder(true).SetTitle("Save Data").SetTitleAlign(tview.AlignLeft)
	app.SetRoot(form, true).SetFocus(form)
	lastForm = form
}

// getData retrieves stored data and displays it in a list.
func getData(app *tview.Application, client pb.GophKeeperServiceClient, dataType pb.DataType, actionType uint) {
	items, err := handlers.GetItems(client, dataType)
	if err != nil {
		errorModal(app, fmt.Sprintf("Failed to retrieve data: %v", err))
		return
	}

	list := tview.NewList()
	num := 1
	for _, item := range items {
		itemCopy := item
		decription := item.Metadata
		if decription == "" {
			decription = "Item " + fmt.Sprint(num)
		}
		list.AddItem(decription, fmt.Sprintf("Type: %v", item.DataType), 0, func() {
			showDataDetails(app, client, itemCopy)
		})
		num++
	}

	list.AddItem("Back", "Return to main menu", 'b', func() {
		dataTypeSelection(app, client, actionType)
	})

	list.SetBorder(true).SetTitle("Saved Data").SetTitleAlign(tview.AlignLeft)
	app.SetRoot(list, true).SetFocus(list)
}

// showDataDetails displays a modal with the selected item's details.
func showDataDetails(app *tview.Application, client pb.GophKeeperServiceClient, item *pb.DataItem) {
	dataContent := string(item.Data)

	modal := tview.NewModal().
		SetText(fmt.Sprintf("Description: %s\n\nData:\n%s", item.Metadata, dataContent)).
		AddButtons([]string{"Back"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			getData(app, client, item.DataType, actions["get"])
		})

	modal.SetBorder(true).SetTitle("Data Details").SetTitleAlign(tview.AlignLeft)
	app.SetRoot(modal, true).SetFocus(modal)
}

// errorModal displays an error message in a modal.
func errorModal(app *tview.Application, message string) {
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"OK"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			app.SetRoot(lastForm, true)
		})

	modal.SetBorder(true).SetTitle("Error").SetTitleAlign(tview.AlignLeft)
	app.SetRoot(modal, true)
}

func promptForFilePath() (string, error) {
	prompt := promptui.Prompt{
		Label: "Enter file path",
		Validate: func(input string) error {
			if _, err := os.Stat(input); os.IsNotExist(err) {
				return fmt.Errorf("file does not exist")
			}
			return nil
		},
	}
	return prompt.Run()
}

func readBinaryFile(filePath string) ([]byte, error) {
	if filePath == "" {
		return nil, fmt.Errorf("file path is empty")
	}
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}
