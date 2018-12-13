package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

const (
	defaultRetries             = 5
	defaultRole                = "demo"
	defaultTimeoutMs           = 2000
	defaultKubernetesTokenPath = "/var/run/secrets/kubernetes.io/serviceaccount"
)

var (
	authMethod    string
	authToken     string
	retries       int
	role          string
	timeout       time.Duration
	tokenPath     string
	vaultAuthPath string
	vaultAddr     string
	vaultPrefix   string
)

func init() {
	pflag.StringVar(&authMethod, "auth-method", "", "auth method to use for authenitcation")
	pflag.StringVar(&authToken, "auth-token", "", "auth token to use with vault")
	pflag.IntVar(&retries, "retries", defaultRetries, "number of retries")
	pflag.StringVar(&role, "k8s-role", defaultRole, "k8s role to authentication against vault with")
	pflag.DurationVar(&timeout, "vault-timeout", defaultTimeoutMs, "timeout for vault requests in milliseconds")
	pflag.StringVar(&tokenPath, "token-path", "", "path on the filesystem to find the JWT")
	pflag.StringVar(&vaultAddr, "vault-addr", os.Getenv("VAULT_ADDR"), "address to access vault and should be a full URL")
	pflag.StringVar(&vaultPrefix, "vault-prefix", "", "path in vault to begin looking for secrets")

	pflag.Parse()
}

func main() {

	if vaultPrefix == "" {
		log.Fatalf("a Vault prefix must be specified")
	}
	client, err := vault.NewClient(&vault.Config{
		Address:    vaultAddr,
		MaxRetries: retries,
		Timeout:    timeout * time.Millisecond,
	})
	if err != nil {
		log.Fatalf("unable to connect to vault: %+v\n", err)
	}
	data := make(map[string]interface{})
	fileToken := []byte{}
	err = errors.New("")
	f := &os.File{}
	t := ""

	data["role"] = role

	if tokenPath == "" && authMethod == "kubernetes" {
		tokenPath = defaultKubernetesTokenPath
	}

	if authToken == "" {
		for err != nil {
			f, err = os.Open(tokenPath)
			if err != nil {
				log.Printf("Unable to access secrets file: %+v", err)
				time.Sleep(100 * time.Millisecond)
			}
		}
		defer f.Close()

		fileToken, err = ioutil.ReadAll(f)
		if err != nil {
			log.Fatalf("Unable to read JWT for service account: %+v", err)
		}
	}

	switch authMethod {
	case "kubernetes":
		vaultAuthPath = "auth/kube-uw2-110/login"
		if authToken != "" {
			t = authToken
		} else {
			t = string(fileToken)
		}
		data["jwt"] = t
	case "github":
		vaultAuthPath = "auth/github/login"
		if authToken != "" {
			t = authToken
		} else {
			t = string(fileToken)
		}
		data["token"] = t
	default:
		log.Fatalf("auth method %s not implemented", authMethod)
	}

	secret, err := client.Logical().Write(vaultAuthPath, data)
	if err != nil {
		log.Fatalf("unable to login to vault on %s: %+v", vaultAuthPath, err)
	}

	token, err := secret.TokenID()
	if err != nil {
		log.Fatalf("unable to lookup token: %+v", err)
	}
	client.SetToken(token)

	vAuth := client.Auth().Token()
	defer vAuth.RevokeSelf("")

	if err := getSecrets(client, vaultPrefix, ""); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

}

func getSecrets(client *vault.Client, vPath, postfix string) error {
	secrets, err := client.Logical().List(path.Join(vPath, postfix))
	if err != nil {
		return errors.Wrap(err, "unable to list path"+path.Join(vPath, postfix))
	}

	if secrets == nil || secrets.Data == nil {
		createEnvVar(client, vPath, postfix)
		return nil
	}

	for name, secret := range secrets.Data {
		sval, ok := secret.([]interface{})
		if !ok {
			return fmt.Errorf("secret is unexpected type: %T", secret)
		}
		for _, s := range sval {
			if name == "keys" && s != nil {
				err := getSecrets(client, vPath, path.Join(postfix, s.(string)))
				if err != nil {
					return errors.Wrap(err, "unable to get secret"+path.Join(vPath, postfix, s.(string)))
				}
			}
		}
	}

	return nil
}

func createEnvVar(client *vault.Client, vPath, postfix string) error {
	secret, err := client.Logical().Read(path.Join(vPath, postfix))
	if err != nil {
		return err
	}

	if secret == nil {
		fmt.Printf("# warning, value for secret at %s not found! skipping...\n", path.Join(vPath, postfix))
		return nil
	}

	if v, ok := secret.Data["value"]; ok {
		envVar := fmt.Sprintf("%s", strings.ToUpper(strings.Replace(postfix, "/", "_", -1)))
		fmt.Printf("export %s=%s\n", envVar, v.(string))
	}

	return nil
}
