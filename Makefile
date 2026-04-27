BIN_DIR = bin
TARGET = $(BIN_DIR)/bookstore-manager
ADMIN_TARGET = $(BIN_DIR)/admin-manager

SRC = ./cmd/bookstore-manager/bookstore-manager.go
ADMIN_SRC = ./cmd/admin-manager/admin-manager.go

bookstore-manager:
	go build -o $(TARGET) $(SRC)

admin-manager:
	go build -o $(ADMIN_TARGET) $(ADMIN_SRC)

clean:
	rm -rf $(BIN_DIR)

run-bookstore-manager:
	$(TARGET)

run-admin-manager:
	$(ADMIN_TARGET)
