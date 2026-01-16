package main

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"
)

const (
	max_mem = 2 * 1024 * 1024
	folder  = "./enc/"
)

func generateSecureRandomString(length int) ([]byte, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return []byte{}, err
	}
	for i := range b {
		b[i] = charset[int(b[i])%len(charset)]
	}
	return b, nil
}

func post_file(w http.ResponseWriter, r *http.Request) {
	log.Printf("POST /file - %s\n", r.RemoteAddr)
	r.ParseMultipartForm(max_mem)
	var buff bytes.Buffer

	file, header, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Failed to receive file: %s", err.Error())
		return
	}
	defer file.Close()

	fmt.Printf("Filename: [%s]\n", header.Filename)
	io.Copy(&buff, file)
	fmt.Printf("%d bytes\n", len(buff.Bytes()))

	id := uuid.New()

	key, err := generateSecureRandomString(32)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to generate secure string: %s", err.Error())
		return
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to generate cipher: %s", err.Error())
		return
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to make GCM: %s", err.Error())
		return
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to make nonce: %s", err.Error())
		return
	}

	ciphertext := aesGCM.Seal(nonce, nonce, buff.Bytes(), nil)

	err = os.WriteFile(folder+id.String(), ciphertext, 0600)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to write encrypted data to disk: %s", err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprintf(w, "\"https://%s/?id=%s&key=%s\"\n", r.Host, id.String(), key)
}

func get_file(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET /file - %s\n", r.RemoteAddr)
	q := r.URL.Query()

	id_str, err := uuid.Parse(q.Get("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid ID: %s", err.Error())
		return
	}

	key := q.Get("key")

	ciphertext, err := os.ReadFile(folder + id_str.String())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to read encrypted data from disk: %s", err.Error())
		return
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to make block from cipher key: %s", err.Error())
		return
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to make aesGCM: %s", err.Error())
		return
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Cipher text length less than nonce size")
		return
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to construct plaintext: %s", err.Error())
		return
	}

	w.WriteHeader(200)
	w.Write(plaintext)
}

func delete_file(w http.ResponseWriter, r *http.Request) {
	log.Printf("DEL /file - %s\n", r.RemoteAddr)
	q := r.URL.Query()

	id_str, err := uuid.Parse(q.Get("id"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid ID: %s", err.Error())
		return
	}

	err = os.Remove(folder + id_str.String())
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprintf(w, "Failed to remove file: %s", err.Error())
		return
	}

	w.WriteHeader(http.StatusGone)
	fmt.Fprintf(w, "Deleted %s\n", id_str.String())
}
