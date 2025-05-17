# Credentials PDF generator

This submodule generates a PDF file with credentials obtained from a [`players`](../players/) state.

To generate the file, run the following.

```bash
(cd ../players && pulumi stack output players --show-secrets) | go run main.go
```
