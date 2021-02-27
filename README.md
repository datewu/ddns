# Go Project-abc

## go project template for github.

### Features
1. Makefile
2. Dockerfile
3. Version.go
4. github ACTIONS (you may need change Github secrets settings)

### Preference
1. LOG: github.com/rs/zerolog
2. TESTING: github.com/stretchr/testify

### ps

Remember to run `go mod init` after git clone:

```bash
git clone github.com/datewu/project-abc
## change Makefile #L59: @go mod init github.com/YOUR_REPO/${APP}
make custom

## edit github secrets settings
``` 