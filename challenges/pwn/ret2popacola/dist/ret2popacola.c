
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <time.h>

#define BUFF_SIZE 0x100


void SuperSecretAdminAccess() {
    char flag_buffer[256];
    FILE *flag = fopen("./flag.txt", "r");
    if (flag == NULL) {
        printf("Erreur : impossible de vous afficher les instructions.\n");
        return;
    }
    printf("Bienvenue cher administrateur. Le plan de contre attaque contre FreizhCola est en cours.\n");
    printf("Voici les instructions : \n");\

    while (fgets(flag_buffer, sizeof(flag_buffer), flag) != NULL) {
        printf("%s", flag_buffer);
    }
    fflush(stdout);
    fclose(flag);
}

void sellTracker() {
    char name[64];

    printf("Bonjour et bienvenue sur l'outil de consultation des ventes de PopaCola !\n");
    printf("Veuillez rentrer votre login pour acceder aux ventes du jour.\n");
    fflush(stdout);

    fgets(name, BUFF_SIZE, stdin);
    if (strcmp(name, "popacola\n") == 0) {
        srand(time(NULL));
        int numberPopaCola = rand() % 1000;
        int numberFreizhCola = rand() % 100;
        printf("Ravi de vous revoir ! Aujourd'hui, %d PopaCola ont ete vendues.\n", numberPopaCola);
        printf("C'est %d plus que FreizhCola, encore une belle victoire.\n", numberPopaCola - numberFreizhCola);
    } else {
        printf("Vous n'etes pas autorise a voir les chiffres de vente !\n");
    }
        
}

int main() {
    sellTracker();
    return 0;
}
