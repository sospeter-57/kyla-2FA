package main

import (
	"fmt"
	"kyla-2FA/utils"
	"log"
	"time"

	"github.com/pquerna/otp/totp"
)

func main() {
	randomBytes := utils.GenerateRandomBytes(32)
	encodedSecret := utils.EncodeBase64(randomBytes)

	code, err := totp.GenerateCode(encodedSecret, time.Now())
	if err != nil {
		log.Fatal("Error: ", err)
	}

	fmt.Println("TOTP code: ", code)
}
