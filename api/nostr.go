package api

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/nbd-wtf/go-nostr"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
	"time"
)

const keyFile = "nostr.key"

func createKeyFile(privateKey string) error {
	return os.WriteFile(keyFile, []byte(privateKey), 0644)
}

func getPrivateKey() (*string, error) {
	if _, err := os.Stat(keyFile); errors.Is(err, os.ErrNotExist) {
		privateKey := nostr.GeneratePrivateKey()
		err = createKeyFile(privateKey)
		if err != nil {
			return nil, err
		}
		return &privateKey, nil
	}
	dat, err := os.ReadFile(keyFile)
	if err != nil {
		return nil, err
	}
	privateKey := string(dat)
	return &privateKey, nil

}

func ConnectNostr(pool *nostr.RelayPool, relays []string) error {
	privateKey, err := getPrivateKey()
	if err != nil {
		return err
	}
	fmt.Println(nostr.GetPublicKey(*privateKey))
	pool.SecretKey = privateKey
	for _, r := range relays {
		err = <-pool.Add(r, nostr.SimplePolicy{Read: true, Write: true})
		if err != nil {
			return err
		}
	}
	return nil
}

func GetSubscriptionFilter(toPublicKey string) nostr.Filters {
	t := time.Now()
	subscriptionsTags := make(nostr.TagMap, 0)
	if toPublicKey != "" {
		subscriptionsTags["p"] = nostr.Tag{toPublicKey}
	}
	return nostr.Filters{
		{
			Tags:  subscriptionsTags,
			Kinds: []int{nostr.KindEncryptedDirectMessage},
			Since: &t,
		}}
}

func SubscribeNostrEvents(pool *nostr.RelayPool, filter nostr.Filters, callback func(message nostr.Event)) {
	fromPublicKey, err := nostr.GetPublicKey(*pool.SecretKey)
	if err != nil {
		panic(err)
	}
	sub, events := pool.Sub(filter)
	log.WithField("subscription", sub).Infof("started nostr subscription")
	go func() {
		for event := range nostr.Unique(events) {
			if event.Tags.ContainsAny("p", []string{fromPublicKey}) {
				callback(event)
			}
		}
	}()

}
func PublishNostrEvents(content, publicKey string, pool *nostr.RelayPool) {
	secret, err := ComputeSharedSecret(*pool.SecretKey, publicKey)
	if err != nil {
		panic(err)
	}
	tags := make(nostr.Tags, 0)
	tags = append(tags, nostr.Tag{"p", publicKey})
	msg, err := Encrypt(content, secret)
	event, statuses, err := pool.PublishEvent(&nostr.Event{
		CreatedAt: time.Now(),
		Kind:      nostr.KindEncryptedDirectMessage,
		Tags:      tags,

		Content: msg,
	})
	if err != nil {
		fmt.Printf("error calling PublishEvent(): %s\n", err.Error())
	}
	StatusProcess(event, statuses)
}

// handle events from out publish events
func StatusProcess(event *nostr.Event, statuses chan nostr.PublishStatus) {
	for status := range statuses {
		switch status.Status {
		case nostr.PublishStatusSent:
			log.WithField("relay", status.Relay).WithField("id", event.ID).Infof("event succesfully sent")
			return
		case nostr.PublishStatusFailed:
			log.WithField("relay", status.Relay).WithField("id", event.ID).Errorf("failed to send event")
			return
		case nostr.PublishStatusSucceeded:
			log.WithField("relay", status.Relay).WithField("id", event.ID).Infof("event succesfully seen")
			return
		}
	}
}

// aes-256-cbc
func Encrypt(message string, key []byte) (string, error) {
	// block size is 16 bytes
	iv := make([]byte, 16)
	// can probably use a less expensive lib since IV has to only be unique; not perfectly random; math/rand?
	_, err := rand.Read(iv)
	if err != nil {
		return "", fmt.Errorf("Error creating initization vector: %s. \n", err.Error())
	}

	// automatically picks aes-256 based on key length (32 bytes)
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("Error creating block cipher: %s. \n", err.Error())
	}
	mode := cipher.NewCBCEncrypter(block, iv)

	// PKCS5 padding
	padding := block.BlockSize() - len([]byte(message))%block.BlockSize()
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	paddedMsgBytes := append([]byte(message), padtext...)

	ciphertext := make([]byte, len(paddedMsgBytes))
	mode.CryptBlocks(ciphertext, paddedMsgBytes)

	return base64.StdEncoding.EncodeToString(ciphertext) + "?iv=" + base64.StdEncoding.EncodeToString(iv), nil
}

// ECDH
func ComputeSharedSecret(senderPrivKey string, receiverPubKey string) (sharedSecret []byte, err error) {
	privKeyBytes, err := hex.DecodeString(senderPrivKey)
	if err != nil {
		return nil, fmt.Errorf("Error decoding sender private key: %s. \n", err)
	}
	privKey, _ := btcec.PrivKeyFromBytes(privKeyBytes)

	// adding 02 to signal that this is a compressed public key (33 bytes)
	pubKeyBytes, err := hex.DecodeString("02" + receiverPubKey)
	if err != nil {
		return nil, fmt.Errorf("Error decoding hex string of receiver public key: %s. \n", err)
	}
	pubKey, err := btcec.ParsePubKey(pubKeyBytes)
	if err != nil {
		return nil, fmt.Errorf("Error parsing receiver public key: %s. \n", err)
	}

	return btcec.GenerateSharedSecret(privKey, pubKey), nil
}

// aes-256-cbc
func Decrypt(content string, key []byte) (string, error) {
	parts := strings.Split(content, "?iv=")
	if len(parts) < 2 {
		return "", fmt.Errorf("Error parsing encrypted message: no initilization vector. \n")
	}

	ciphertext, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("Error decoding ciphertext from base64: %s. \n", err)
	}

	iv, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("Error decoding iv from base64: %s. \n", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("Error creating block cipher: %s. \n", err.Error())
	}
	mode := cipher.NewCBCDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	mode.CryptBlocks(plaintext, ciphertext)

	return string(plaintext[:]), nil
}
