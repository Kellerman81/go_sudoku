BINARY   = go_sudoku.exe
LDFLAGS  = -H windowsgui

.PHONY: build run build-linux build-android clean

# Windows (no console window)
build:
	go build -ldflags="$(LDFLAGS)" -o $(BINARY) .

run:
	go run -ldflags="$(LDFLAGS)" .

# Linux desktop
build-linux:
	GOOS=linux GOARCH=amd64 go build -o go_sudoku .

# Android (install gogio first: go install gioui.org/cmd/gogio@latest)
install-gogio:
	go install gioui.org/cmd/gogio@latest

build-android:
	gogio -target android -o go_sudoku.apk .

clean:
	del /f $(BINARY) 2>nul || true
