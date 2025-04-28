#include <windows.h>
#include <wincrypt.h>
#include <commdlg.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define MAX_BUF 8192

void get_password(char *buf, DWORD size) {
    HANDLE hStdin = GetStdHandle(STD_INPUT_HANDLE);
    DWORD mode;
    GetConsoleMode(hStdin, &mode);
    SetConsoleMode(hStdin, mode & ~ENABLE_ECHO_INPUT);
    printf("Master password: ");
    fgets(buf, size, stdin);
    SetConsoleMode(hStdin, mode);
    printf("\n");
    buf[strcspn(buf, "\n")] = 0;
}

void show_menu() {
    printf("\n== Password Manager ==\n");
    printf("[1] View entries\n");
    printf("[2] Add entry\n");
    printf("[3] Delete entry\n");
    printf("[0] Save & Quit\n");
    printf("Choice: ");
}

void xor_buffer(BYTE *buf, DWORD buflen, const char *key, DWORD keylen) {
    for (DWORD i = 0; i < buflen; i++) {
        buf[i] ^= key[i % keylen];
    }
}

char *base64_encode(const BYTE *data, DWORD datalen, DWORD *outlen) {
    CryptBinaryToStringA(data, datalen, CRYPT_STRING_BASE64 | CRYPT_STRING_NOCRLF, NULL, outlen);
    char *out = malloc(*outlen);
    CryptBinaryToStringA(data, datalen, CRYPT_STRING_BASE64 | CRYPT_STRING_NOCRLF, out, outlen);
    return out;
}

int main() {
    OPENFILENAME ofn;
    char filename[MAX_PATH] = {0};
    ZeroMemory(&ofn, sizeof(ofn));
    ofn.lStructSize = sizeof(ofn);
    ofn.lpstrFilter = "PassDB\0*.passdb\0";
    ofn.lpstrFile = filename;
    ofn.nMaxFile = MAX_PATH;
    ofn.Flags = OFN_FILEMUSTEXIST | OFN_PATHMUSTEXIST;
    if (!GetOpenFileName(&ofn)) return 1;

    HANDLE hFile = CreateFile(filename, GENERIC_READ | GENERIC_WRITE, 0, NULL, OPEN_EXISTING, 0, NULL);
    if (hFile == INVALID_HANDLE_VALUE) return 1;

    DWORD size = GetFileSize(hFile, NULL), read;
    BYTE *buf = malloc(MAX_BUF);
    ZeroMemory(buf, MAX_BUF);
    ReadFile(hFile, buf, size, &read, NULL);

    char pwd[128];
    get_password(pwd, sizeof(pwd));

    DWORD b64len;
    char *b64key = base64_encode((BYTE*)pwd, strlen(pwd), &b64len);

    xor_buffer(buf, read, b64key, b64len);
    buf[read] = 0;

    while (1) {
        show_menu();
        int choice;
        scanf("%d", &choice); getchar();
        if (choice == 0) break;
        else if (choice == 1) {
            printf("\n--- Entries ---\n%s", buf);
        } else if (choice == 2) {
            char label[64], user[64], pass[64];
            printf("Label: "); fgets(label, 64, stdin); label[strcspn(label, "\n")] = 0;
            printf("Username: "); fgets(user, 64, stdin); user[strcspn(user, "\n")] = 0;
            printf("Password: "); fgets(pass, 64, stdin); pass[strcspn(pass, "\n")] = 0;
            char entry[256];
            snprintf(entry, sizeof(entry), "%s|%s|%s\n", label, user, pass);
            strncat((char*)buf, entry, MAX_BUF - strlen((char*)buf) - 1);
        } else if (choice == 3) {
            char label[64];
            printf("Label to delete: "); fgets(label, 64, stdin); label[strcspn(label, "\n")] = 0;
            char *line = strtok((char*)buf, "\n");
            char temp[MAX_BUF] = "";
            while (line) {
                if (strncmp(line, label, strlen(label)) != 0 || line[strlen(label)] != '|') {
                    strncat(temp, line, MAX_BUF - strlen(temp) - 2);
                    strncat(temp, "\n", MAX_BUF - strlen(temp) - 1);
                }
                line = strtok(NULL, "\n");
            }
            strncpy((char*)buf, temp, MAX_BUF);
        }
    }

    SetFilePointer(hFile, 0, NULL, FILE_BEGIN);
    DWORD dataLen = (DWORD)strlen((char*)buf);
    xor_buffer(buf, dataLen, b64key, b64len);
    DWORD written;
    WriteFile(hFile, buf, dataLen, &written, NULL);
    SetEndOfFile(hFile);

    CloseHandle(hFile);
    free(buf);
    free(b64key);
    return 0;
}
