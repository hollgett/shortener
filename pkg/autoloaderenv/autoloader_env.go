package autoloaderenv

import (
	"bytes"
	"os"
)

const fileNameEnv = ".env"

// LoadEnv set env.
//
// opening default file ".env", parse and set env.
func LoadEnv() {
	src, err := getFileData()
	if err != nil {
		return
	}
	setEnv(parseByte(src))
}

func getFileData() ([]byte, error) {
	src, err := os.ReadFile(fileNameEnv)
	if err != nil {
		return nil, err
	}
	return src, nil
}

func setEnv(envs map[string]string) {
	if len(envs) == 0 {
		return
	}
	for k, v := range envs {
		os.Setenv(k, v)
	}
}

func parseByte(src []byte) map[string]string {
	buf := bytes.Split(bytes.ReplaceAll(src, []byte("\r\n"), []byte("\n")), []byte("\n"))
	envs := make(map[string]string)
	for _, v := range buf {
		if len(v) == 0 || v[0] == '#' {
			continue
		}
		b := bytes.SplitN(v, []byte("="), 2)
		if len(b) == 2 {
			envs[string(b[0])] = string(b[1])
		} else {
			envs[string(b[0])] = ""
		}

	}
	return envs
}
