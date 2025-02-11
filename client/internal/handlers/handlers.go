// Package handlers provides functions to manage authentication, data encryption,
// and communication with the GophKeeper gRPC service.
package handlers

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	pb "github.com/golangTroshin/gophkeeper/grpc"
	"github.com/rivo/tview"
	"golang.org/x/crypto/pbkdf2"
)

// Session stores the user authentication token.
type Session struct {
	UserToken string
}

// Global session instance.
var session = &Session{}

// Login authenticates a user and retrieves a session token.
func Login(client pb.GophKeeperServiceClient, username, password string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.AuthenticateUser(ctx, &pb.AuthenticateUserRequest{
		Username: username,
		Password: password,
	})

	if err != nil {
		return err
	}

	if !res.Success {
		return fmt.Errorf("%s", res.Message)
	}

	session.UserToken = res.Token
	return nil
}

// SignUp registers a new user and saves their master seed for encryption.
func SignUp(client pb.GophKeeperServiceClient, username, password, seed string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := client.RegisterUser(ctx, &pb.RegisterUserRequest{
		Username: username,
		Password: password,
		Seed:     seed,
	})

	if err != nil {
		return err
	}

	if !res.Success {
		return fmt.Errorf("%s", res.Message)
	}

	session.UserToken = res.Token
	return nil
}

// CollectFormData retrieves user input from the form for different data types.
func CollectFormData(form *tview.Form, dataType pb.DataType) map[string]string {
	data := make(map[string]string)

	switch dataType {
	case pb.DataType_CREDENTIALS:
		data["login"] = form.GetFormItemByLabel("Login").(*tview.InputField).GetText()
		data["password"] = form.GetFormItemByLabel("Password").(*tview.InputField).GetText()
	case pb.DataType_TEXT:
		data["text"] = form.GetFormItemByLabel("Text").(*tview.InputField).GetText()
	case pb.DataType_BINARY:
		data["file_path"] = form.GetFormItemByLabel("Selected File").(*tview.InputField).GetText()
	case pb.DataType_CARD:
		data["card_number"] = form.GetFormItemByLabel("Card Number").(*tview.InputField).GetText()
		data["expiration_date"] = form.GetFormItemByLabel("Expiration Date").(*tview.InputField).GetText()
		data["cvv"] = form.GetFormItemByLabel("CVV").(*tview.InputField).GetText()
	}
	data["metadata"] = form.GetFormItemByLabel("Description").(*tview.InputField).GetText()
	return data
}

// SaveData encrypts user data and sends it to the server for storage.
func SaveData(client pb.GophKeeperServiceClient, app *tview.Application, dataType pb.DataType, data map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if session.UserToken == "" {
		return fmt.Errorf("user is not authenticated")
	}

	metaData := data["metadata"]
	delete(data, "metadata")

	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	res, err := client.MasterSeedRetrieve(ctx, &pb.MasterSeedRetrieveRequest{Token: session.UserToken})
	if err != nil || !res.Success {
		return err
	}
	key := DeriveKeyFromSeed(string(res.MasterSeed))
	encryptedData, err := encryptData(bytes, key)
	if err != nil {
		return err
	}

	var resp *pb.StoreDataResponse

	resp, err = client.StoreData(ctx, &pb.StoreDataRequest{
		Token:    session.UserToken,
		DataType: dataType,
		Data:     encryptedData,
		Metadata: metaData,
	})

	if err != nil {
		return err
	}

	if !res.Success {
		return fmt.Errorf("failed to save data: %v ||| %v", err, resp.Message)
	}

	return nil
}

// DeriveKeyFromSeed generates a cryptographic key using PBKDF2.
func DeriveKeyFromSeed(seed string) []byte {
	salt := []byte("LOnhFQ:zixsQ")
	return pbkdf2.Key([]byte(seed), salt, 4096, 32, sha256.New)
}

// encryptData encrypts plaintext data using AES-GCM.
func encryptData(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return make([]byte, 0), err
	}

	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return make([]byte, 0), err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return make([]byte, 0), err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// DecryptData decrypts ciphertext using AES-GCM.
func DecryptData(encryptedText string, key []byte) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	if len(data) < 12 {
		return "", errors.New("invalid ciphertext")
	}

	nonce, ciphertext := data[:12], data[12:]

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// GetItems retrieves encrypted data from the server, decrypts it, and returns the items.
func GetItems(client pb.GophKeeperServiceClient, dataType pb.DataType) ([]*pb.DataItem, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	items := []*pb.DataItem{}

	res, err := client.RetrieveData(ctx, &pb.RetrieveDataRequest{
		Token:  session.UserToken,
		Filter: dataType,
	})

	if err != nil {
		return items, err
	}

	var resSeed *pb.MasterSeedRetrieveResponse
	resSeed, err = client.MasterSeedRetrieve(ctx, &pb.MasterSeedRetrieveRequest{Token: session.UserToken})
	if err != nil || !resSeed.Success {
		return items, err
	}

	key := DeriveKeyFromSeed(resSeed.MasterSeed)

	for _, item := range res.Items {
		decryptedData, err := DecryptData(base64.StdEncoding.EncodeToString(item.Data), key)
		if err != nil {
			return items, err
		}

		item.Data = []byte(decryptedData)
	}

	return res.Items, nil
}
