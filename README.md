**kyla-2FA**

kyla-2FA is a cross-platform, desktop two-factor authentication (2FA) application built with Go and Fyne. It generates **TOTP*** codes for multiple accounts, helping users secure their online services with industry-standard two-factor authentication.

**Features**

- Multi-account support for multiple sites/users

- Secure local storage of secrets (encrypted)

- Manual secret entry for sites without QR codes

- Copy TOTP codes to clipboard

- Auto-refresh of TOTP codes with 30-second countdown

- PIN or biometric lock for app access


**Installation**

1. Ensure you have Go 1.26+ installed.

2. Clone the repository:
```
git clone https://github.com/sospeter-57/kyla-2FA.git
cd kyla-2FA
```
3. Install dependencies:
```
go get fyne.io/fyne/v2
go get fyne.io/fyne/v2/app
go get fyne.io.fyne/v2/canvas
go get fyne.io/fyne/v2/layout
go get fyne.io/fyne/v2/widget
go get fyne.io/fyne/v2/container
go get github.com/pquerna/otp
```
**Usage**

Run the application:
```
go run main.go
```
On first launch, add accounts by scanning a QR code or entering the secret manually.

Select the account/site from the dropdown to generate TOTP codes.

Use the copy button to quickly copy codes to your clipboard.

**Roadmap / Future Enhancements**
- Dark/light theme support

- QR code scanning for easy account setup

- Backup and restore of all accounts

- Implement encrypted cloud sync for cross-device usage.

- Add biometric authentication for app access (fingerprint/face).

- Support notifications for expiring codes.

- Cross-device sync with encrypted cloud storage (optional)

- Secure export/import of accounts

- Optional notifications/reminders for code refresh

**Contributing**

Contributions are welcome! Please fork the repo, make your changes, and submit a pull request. Ensure all secrets are encrypted before committing.

**License**
Apache 2.0 LICENSE: https://www.apache.org/licenses/LICENSE-2.0


**Building from source**

building for linux
    go build -o kyla-2FA main.go

building for windows
    env CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o kyla-2FA.exe main.go
    dependencies
        -
