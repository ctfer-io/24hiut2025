BUILD_DIR 	= build
OBJ_NAME	= $(BUILD_DIR)/freizh-exam.o
BIN_NAME	= $(BUILD_DIR)/freizh-exam

SOURCES 	= src/main.asm

all: clean package


clean:
	@echo -n Clean build dir...
	@rm -rf $(BUILD_DIR)/
	@echo OK

package: build
	zip -j $(BUILD_DIR)/freizh-exam.zip $(BUILD_DIR)/freizh-exam

build:
	@echo -n Build bin...
	@mkdir $(BUILD_DIR)/
	@nasm -f elf64 $(SOURCES) -o $(OBJ_NAME)
	@ld $(OBJ_NAME) -o $(BIN_NAME)
	@strip -s $(BIN_NAME)
	@echo OK
