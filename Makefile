BIN_DIR = bin
TARGET = $(BIN_DIR)\bookstore-manager.exe
ADMIN_TARGET = $(BIN_DIR)\admin-manager.exe

SRC = .\cmd\bookstore-manager\bookstore-manager.go
ADMIN_SRC = .\cmd\admin-manager\admin-manager.go

bookstore-manager:
	go build -o $(TARGET) $(SRC)

admin-manager:
	go build -o $(ADMIN_TARGET) $(ADMIN_SRC)

clean:
	if exist $(BIN_DIR) rmdir /s /q $(BIN_DIR)

run bookstore-manager:
	$(TARGET)

run admin-manager:
	$(ADMIN_TARGET)