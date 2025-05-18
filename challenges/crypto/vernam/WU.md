# Write-Up - Crypto / Mr Vernam

Dans ce challenge, l'objectif est de faire face à l'algorithme de chiffrement de Vernam, réputé parfaitement sûr.
Il requiert 3 conditions pour cela:
- la clé est aussi grande que le message
- la clé est parfaitement aléatoire (on le suppose)
- la clé n'est pas réutilisée (on le suppose)

Dans la mise en oeuvre, toutefois, il semblerait que la clé ne soit pas aussi grande que le message (c'est la seule chose qu'on peut supposer défaillante dans la mise en oeuvre).
Cela va amener à une réutilisation de ladite clé au cours du chiffrement, réduisant le caractère aléatoire du chiffrement.
De plus, on sait que le chiffré est sous la forme `24HIUT{...}`, donc nous disposons déjà d'une partie du message.

Nous allons donc mener une attaque crypto KPA (Known-Plaintext Attack/Attempt).

## Approche

On sait que le chiffré est encodé en hexa comme `a5bff6b2c2dfc5b8b0e...`.
De plus, on sait que les premiers caractères seront `24HIUT{`.

Ainsi, on a l'égalité suivante.
```
    cipher[0]   = crib[0] ^ key[0]
<=> 0xa5        = '2'     ^ key[0]
<=> 0xa5 ^ '2'  = '2'     ^ key[0] ^ '2'
<=> 0xa5 ^ 0x32 = key[0]
<=> key[0] = 0x97
```

On va pouvoir procéder pas à pas pour la suite.
En formulant l'hypothèse que la clé est plus courte que la partie de clair que l'on connait, on va pouvoir brute-force la clé efficacement.

Ne connaissant pas la taille exacte de la clé, on va ajouter chaque caractère comme précédemment, les uns après les autres.
Notre condition d'arrêt sera dès lors que le clair commencera par `24HIUT` et terminera par `}`.

On code tout cela, comme dans l'exemple qui suit.
```go
package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

func main() {
	cipher, _ := hex.DecodeString("a5bff6b2c2dfc5b8b0eecd8fb7e7d189e4facbdcf8e59e95f2abdd93f2f9dd93f2abce9ae4abd29ab7e8d29ef1abcf8eb0e4d0dbfbea9e8fe5e4cb8df2f6")

	// We know the crib begins by 24HIUT{ as it is the flag format.
	// We can hypothesize that the known part of the crib is larger that the key
	// and verify this.
	// As we also know the last char will be } thanks to flag format,
	// we will determine the proper key based on this one.
	// There is a probability of false positive, but could be manually skipped
	// if necessary.
	key := []byte{}
	base := "24HIUT{"
	for i := range base {
		// Add a known char to the key, one by one
		key = append(key, cipher[i]^base[i])

		// Build the candidate crib
		out := make([]byte, len(cipher))
		for i := range cipher {
			out[i] = cipher[i] ^ key[i%len(key)]
		}
		outStr := string(out)

		// Check if matches conditions (prefix and suffix)
		if strings.HasPrefix(outStr, base) && strings.HasSuffix(outStr, "}") {
			fmt.Printf("Flag: %s ; Key: %s\n", outStr, hex.EncodeToString(key))
			os.Exit(0)
		}
	}
	fmt.Println("Flag not found")
	os.Exit(1)
}
```

On l'exécute, et tout de suite nous disposons du flag.

```bash
$ go run main.go
Flag: 24HIUT{C'est lorsqu'on ne cherche pas la clef qu'on la trouve} ; Key: 978bbefb
```

Ce calcul est automatisé dans ce write-up, mais pourrait être fait manuellement compte tenue de la taille de clé ridicule utilisée.
