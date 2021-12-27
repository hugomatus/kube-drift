package utils

import (
	"crypto/md5"
	"encoding/hex"
	flag "github.com/spf13/pflag"
	"hash/fnv"
	"k8s.io/client-go/util/homedir"
	"os"
	"path/filepath"
	"time"
)

func GetEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		value = fallback
	}
	return value
}

func GetKubeconfig() *string {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	return kubeconfig
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func GetUniqueKey() string {
	h := fnv.New64a()
	// Hash of Timestamp
	h.Write([]byte(time.Now().String()))
	key := hex.EncodeToString(h.Sum(nil))
	return key
}
