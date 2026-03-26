package internal

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	appInstance    fyne.App
	MainWindow     fyne.Window
	vault          Vault
	activePIN      string
	accountSelect  *widget.Select
	codeLabel      *widget.Label
	countdownLabel *widget.Label
)

func InitApp() {
	appInstance = app.NewWithID("io.kyla-2fa")
	appInstance.Settings().SetTheme(theme.DarkTheme())
	MainWindow = appInstance.NewWindow("kyla-2FA")
	MainWindow.Resize(fyne.NewSize(760, 520))
	MainWindow.CenterOnScreen()

	logo := canvas.NewImageFromFile("assets/kyla-2FA_logo.png")
	logo.FillMode = canvas.ImageFillContain
	logo.SetMinSize(fyne.NewSize(180, 80))

	accountSelect = widget.NewSelect([]string{}, onAccountChange)
	accountSelect.PlaceHolder = "Add an account first"
	codeLabel = widget.NewLabel("------")
	codeLabel.TextStyle = fyne.TextStyle{Bold: true}
	codeLabel.Alignment = fyne.TextAlignCenter
	countdownLabel = widget.NewLabel("30s")
	countdownLabel.Alignment = fyne.TextAlignCenter

	addBtn := widget.NewButtonWithIcon("Add account", theme.ContentAddIcon(), showAddAccountDialog)
	removeBtn := widget.NewButtonWithIcon("Remove", theme.DeleteIcon(), removeSelectedAccount)
	copyBtn := widget.NewButtonWithIcon("Copy code", theme.ContentCopyIcon(), copyCurrentCode)
	changePINBtn := widget.NewButtonWithIcon("Change PIN", theme.SettingsIcon(), requestPINChange)
	backupBtn := widget.NewButtonWithIcon("Backup JSON", theme.DocumentSaveIcon(), backupDataDialog)
	restoreBtn := widget.NewButtonWithIcon("Restore JSON", theme.FileIcon(), restoreDataDialog)

	leftPane := container.NewVBox(
		logo,
		widget.NewLabelWithStyle("Accounts", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		accountSelect,
		container.NewGridWithColumns(2, addBtn, removeBtn),
		container.NewGridWithColumns(2, copyBtn, changePINBtn),
		container.NewGridWithColumns(2, backupBtn, restoreBtn),
		layout.NewSpacer(),
		widget.NewLabelWithStyle("Current TOTP", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		codeLabel,
		countdownLabel,
	)

	body := container.NewBorder(nil, nil, nil, nil, leftPane)
	MainWindow.SetContent(body)
}

func ExecuteUnlockFlow() {
	path, err := storagePath()
	if err != nil {
		dialog.ShowError(err, MainWindow)
		return
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		promptNewPIN()
		return
	}

	promptExistingPIN(3)
}

func askPassword(title, message string, success func(string)) {
	pinEntry := widget.NewPasswordEntry()
	pinEntry.SetPlaceHolder("PIN")
	content := container.NewVBox(widget.NewLabel(message), pinEntry)
	dlg := dialog.NewCustomConfirm(title, "OK", "Cancel", content, func(ok bool) {
		if !ok {
			return
		}
		success(pinEntry.Text)
	}, MainWindow)
	dlg.Show()
}

func promptNewPIN() {
	var firstPIN string
	askPassword("Set new PIN", "Enter a new PIN (4+ chars):", func(pin string) {
		if len(strings.TrimSpace(pin)) < 4 {
			dialog.ShowError(fmt.Errorf("PIN should be at least 4 characters"), MainWindow)
			promptNewPIN()
			return
		}
		firstPIN = pin
		askPassword("Confirm PIN", "Re-enter new PIN:", func(confirm string) {
			if firstPIN != confirm {
				dialog.ShowError(fmt.Errorf("PINs do not match"), MainWindow)
				promptNewPIN()
				return
			}
			activePIN = firstPIN
			vault = Vault{Accounts: []Account{}}
			if err := saveVault(activePIN); err != nil {
				dialog.ShowError(err, MainWindow)
				os.Exit(1)
			}
			refreshAccountList()
			startCodeTicker()
		})
	})
}

func promptExistingPIN(attempts int) {
	if attempts <= 0 {
		dialog.ShowError(fmt.Errorf("Too many failed attempts, exiting"), MainWindow)
		appInstance.Quit()
		return
	}

	askPassword("Unlock kyla-2FA", "Enter your PIN:", func(pin string) {
		if pin == "" {
			promptExistingPIN(attempts)
			return
		}
		loaded, err := loadVault(pin)
		if err != nil {
			if errors.Is(err, ErrInvalidPIN) {
				dialog.ShowError(fmt.Errorf("Invalid PIN (%d/%d)", 4-attempts, 3), MainWindow)
				promptExistingPIN(attempts - 1)
				return
			}
			dialog.ShowError(err, MainWindow)
			return
		}
		activePIN = pin
		vault = loaded
		refreshAccountList()
		startCodeTicker()
	})
}

func onAccountChange(selection string) {
	if selection == "" {
		codeLabel.SetText("------")
		return
	}
	account := findAccountByName(selection)
	if account == nil {
		codeLabel.SetText("------")
		return
	}
	updateCodeForAccount(*account)
}

func showAddAccountDialog() {
	name := widget.NewEntry()
	issuer := widget.NewEntry()
	secret := widget.NewEntry()
	secret.SetPlaceHolder("Base32 secret, spaces allowed")

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Friendly name", Widget: name},
			{Text: "Issuer", Widget: issuer},
			{Text: "Secret", Widget: secret},
		},
		OnSubmit: func() {
			if strings.TrimSpace(name.Text) == "" || strings.TrimSpace(secret.Text) == "" {
				dialog.ShowError(fmt.Errorf("Name and Secret are required"), MainWindow)
				return
			}
			cleanSecret := normalizeSecret(secret.Text)
			if len(cleanSecret) < 16 {
				dialog.ShowError(fmt.Errorf("Secret seems too short"), MainWindow)
				return
			}
			// Validate secret by trying to generate TOTP
			testAccount := Account{Secret: cleanSecret, Digits: 6, Period: 30, Algorithm: "SHA1"}
			if _, err := getTOTPCode(testAccount); err != nil {
				dialog.ShowError(fmt.Errorf("Invalid secret format: %v", err), MainWindow)
				return
			}
			account := Account{
				ID:        generateAccountID(name.Text, issuer.Text),
				Name:      strings.TrimSpace(name.Text),
				Issuer:    strings.TrimSpace(issuer.Text),
				Secret:    cleanSecret,
				Algorithm: "SHA1",
				Digits:    6,
				Period:    30,
			}
			vault.Accounts = append(vault.Accounts, account)
			if err := saveVault(activePIN); err != nil {
				dialog.ShowError(err, MainWindow)
				return
			}
			refreshAccountList()
			dialog.ShowInformation("Saved", "Account added", MainWindow)
		},
	}
	// form.Resize(fyne.NewSize(500, 600))

	// dialog.ShowCustom("Add Account", "Cancel", form, MainWindow)
	formContainer := container.NewVBox(form);
	
	dlg := dialog.NewCustom("Add Account", "Cancel", formContainer, MainWindow)
	dlg.Resize(fyne.NewSize(500, 300))
	dlg.Show()
}

func removeSelectedAccount() {
	if accountSelect.Selected == "" {
		dialog.ShowInformation("No account", "Select an account first", MainWindow)
		return
	}
	account := findAccountByName(accountSelect.Selected)
	if account == nil {
		dialog.ShowError(fmt.Errorf("selected account not found"), MainWindow)
		return
	}

	dialog.ShowConfirm("Delete account", fmt.Sprintf("Remove '%s' permanently?", account.Name), func(confirmed bool) {
		if !confirmed {
			return
		}
		for i, a := range vault.Accounts {
			if a.ID == account.ID {
				vault.Accounts = append(vault.Accounts[:i], vault.Accounts[i+1:]...)
				break
			}
		}
		if err := saveVault(activePIN); err != nil {
			dialog.ShowError(err, MainWindow)
			return
		}
		refreshAccountList()
	}, MainWindow)
}

func copyCurrentCode() {
	if codeLabel.Text == "------" || codeLabel.Text == "" {
		dialog.ShowInformation("No code", "No TOTP code available yet", MainWindow)
		return
	}
	MainWindow.Clipboard().SetContent(codeLabel.Text)
	dialog.ShowInformation("Copied", "TOTP copied to clipboard", MainWindow)
}

func requestPINChange() {
	oldPinEntry := widget.NewPasswordEntry()
	newPinEntry := widget.NewPasswordEntry()
	confirmPinEntry := widget.NewPasswordEntry()

	form := &widget.Form{
		Items: []*widget.FormItem{
			{Text: "Current PIN", Widget: oldPinEntry},
			{Text: "New PIN", Widget: newPinEntry},
			{Text: "Confirm new PIN", Widget: confirmPinEntry},
		},
		OnSubmit: func() {
			if oldPinEntry.Text == "" || newPinEntry.Text == "" {
				dialog.ShowError(fmt.Errorf("All fields are required"), MainWindow)
				return
			}
			if newPinEntry.Text != confirmPinEntry.Text {
				dialog.ShowError(fmt.Errorf("New PINs do not match"), MainWindow)
				return
			}
			if _, err := loadVault(oldPinEntry.Text); err != nil {
				dialog.ShowError(fmt.Errorf("Current PIN is incorrect"), MainWindow)
				return
			}
			activePIN = newPinEntry.Text
			if err := saveVault(activePIN); err != nil {
				dialog.ShowError(err, MainWindow)
				return
			}
			dialog.ShowInformation("Success", "PIN changed successfully", MainWindow)
		},
	}
	dialog.ShowCustom("Change PIN", "Cancel", form, MainWindow)
}

func backupDataDialog() {
	fileSave := dialog.NewFileSave(func(u fyne.URIWriteCloser, _ error) {
		if u == nil {
			return
		}
		if err := backupToFile(u.URI().Path()); err != nil {
			dialog.ShowError(err, MainWindow)
			return
		}
		_ = u.Close()
		dialog.ShowInformation("Backup", "Vault backed up successfully", MainWindow)
	}, MainWindow)
	fileSave.SetFileName("kyla-2fa-backup.json")
	fileSave.Show()
}

func restoreDataDialog() {
	dialog.ShowFileOpen(func(r fyne.URIReadCloser, _ error) {
		if r == nil {
			return
		}
		path := r.URI().Path()
		_ = r.Close()
		if err := restoreFromFile(path); err != nil {
			dialog.ShowError(err, MainWindow)
			return
		}
		refreshAccountList()
		dialog.ShowInformation("Restore", "Vault restored and saved", MainWindow)
	}, MainWindow)
}

func refreshAccountList() {
	names := make([]string, 0, len(vault.Accounts))
	for _, account := range vault.Accounts {
		names = append(names, account.Name)
	}
	sort.Strings(names)
	accountSelect.Options = names
	accountSelect.Refresh()
	if len(names) > 0 {
		accountSelect.SetSelected(names[0])
		onAccountChange(names[0])
	} else {
		codeLabel.SetText("------")
		countdownLabel.SetText("--s")
	}
}

func findAccountByName(name string) *Account {
	for i := range vault.Accounts {
		if vault.Accounts[i].Name == name {
			return &vault.Accounts[i]
		}
	}
	return nil
}

func updateCodeForAccount(account Account) {
	code, err := getTOTPCode(account)
	if err != nil {
		codeLabel.SetText("ERR")
		countdownLabel.SetText("--s")
		return
	}
	codeLabel.SetText(code)
}

func startCodeTicker() {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if len(vault.Accounts) == 0 || accountSelect.Selected == "" {
				continue
			}
			acct := findAccountByName(accountSelect.Selected)
			if acct == nil {
				continue
			}
			now := time.Now().Unix()
			secondsLeft := 30 - (now % 30)
			code, _ := getTOTPCode(*acct)
			fyne.Do(func() {
				codeLabel.SetText(code)
				countdownLabel.SetText(fmt.Sprintf("%ds", secondsLeft))
			})
		}
	}()
}

func generateAccountID(name, issuer string) string {
	seed := fmt.Sprintf("%s|%s", name, issuer)
	return fmt.Sprintf("%x", sha256.Sum256([]byte(seed)))
}
