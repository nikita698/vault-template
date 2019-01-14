package main

import (
	"bytes"
	"github.com/Luzifer/rconfig"
	"github.com/Masterminds/sprig"
	"github.com/actano/vault-template/api"
	"io/ioutil"
	"log"
	"os"
	"text/template"
)

var (
	cfg = struct {
		VaultEndpoint  string `flag:"vault,v" env:"VAULT_ADDR" default:"https://127.0.0.1:8200" description:"Vault API endpoint. Also configurable via VAULT_ADDR."`
		VaultTokenFile string `flag:"vault-token-file,f" env:"VAULT_TOKEN_FILE" description:"The file which contains the vault token. Also configurable via VAULT_TOKEN_FILE."`
		TemplateFile   string `flag:"template,t" env:"TEMPLATE_FILE" description:"The template file to render. Also configurable via TEMPLATE_FILE."`
		OutputFile     string `flag:"output,o" env:"OUTPUT_FILE" description:"The output file. Also configurable via OUTPUT_FILE."`
	}{}
)

func usage(msg string) {
	println(msg)
	rconfig.Usage()
	os.Exit(1)
}

func config() {
	rconfig.Parse(&cfg)

	if cfg.VaultTokenFile == "" {
		usage("No vault token file given")
	}

	if cfg.TemplateFile == "" {
		usage("No template file given")
	}

	if cfg.OutputFile == "" {
		usage("No output file given")
	}
}

func renderTemplate(vaultClient api.VaultClient, templateContent string) (*bytes.Buffer, error) {
	funcMap := template.FuncMap{
		"vault": vaultClient.QuerySecret,
	}

	tmpl, err := template.
		New("template").
		Funcs(sprig.TxtFuncMap()).
		Funcs(funcMap).
		Parse(templateContent)

	if err != nil {
		return nil, err
	}

	var outputBuffer bytes.Buffer

	if err := tmpl.Execute(&outputBuffer, nil); err != nil {
		return nil, err
	}

	return &outputBuffer, nil
}

func main() {
	config()

	vaultToken, err := ioutil.ReadFile(cfg.VaultTokenFile)

	if err != nil {
		log.Fatalf("Unable to read vault token file: %s", err)
	}

	vaultClient, err := api.NewVaultClient(cfg.VaultEndpoint, string(vaultToken))

	if err != nil {
		log.Fatalf("Unable to create vault client: %s", err)
	}

	templateContent, err := ioutil.ReadFile(cfg.TemplateFile)

	if err != nil {
		log.Fatalf("Unable to read template file: %s", err)
	}

	outputBuffer, err := renderTemplate(vaultClient, string(templateContent))

	if err != nil {
		log.Fatalf("Unable to render template: %s", err)
	}

	outputFile, err := os.Create(cfg.OutputFile)

	if err != nil {
		log.Fatalf("Unable to write output file: %s", err)
	}

	defer outputFile.Close()

	outputFile.Write(outputBuffer.Bytes())
}
