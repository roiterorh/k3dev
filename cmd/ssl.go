package cmd

import (
	"github.com/zain-bahsarat/minica"
	"os"
	// "os/exec"
	"github.com/apex/log/handlers/cli"
	"github.com/apex/log"
)
func init() {
	log.SetLevel(log.DebugLevel)
	log.SetHandler(cli.Default)
}
func createCertificates(domain string) {

	ctx := log.WithFields(log.Fields{
		"domain": domain,
	})


	cert_path := UserHomeDir() + "/.k3dev/certificates"
	if err := os.MkdirAll(cert_path, os.ModePerm); err != nil {
		ctx.Fatalf("could not create folder: %v", err)

	}
	
	issuer, err := minica.GetIssuer(cert_path+"/ca.key", cert_path+"/ca.crt")

	if err != nil {
		ctx.Fatalf("could not issue ca: %v", err)
	}

	os.Remove(cert_path + "/key.pem")
	os.Remove(cert_path + "/cert.pem")
	_, err = minica.Sign(issuer, []string{"*." + domain}, nil, cert_path)

	if err != nil {
		ctx.Fatalf("could not issue certificates: %v", err)

		}

// 	out, err := exec.Command("certutil", "-d", "/home/nroitero/.pki/nssdb", "-L", "-n", "k3de").Output()
// 	if err.Error()	== "exec: \"certutil\": executable file not found in $PATH" {
// 		ctx.Fatalf("Missing binary: %v", err)
// 		}
// 		if len(out)!=0{
// 					_, err = exec.Command("certutil", "-d", "/home/nroitero/.pki/nssdb", "-D", "-n", "k3dev").Output()
// 		if err != nil {
// 			ctx.Fatalf("Could not remove CA: %v", err)
// 		}
// 		ctx.Debug("CA removed from trust store")

// 		}
// 	_, err = exec.Command("certutil", "-d", "/home/nroitero/.pki/nssdb", "-A", "-n", "k3dev", "-t", "TC,C,c", "-i", "/home/nroitero/.k3dev/certificates/ca.crt").Output()
// if err!=nil{
// 		ctx.Fatalf("Could not add CA to trust store: %v", err)
// }
// ctx.Info("CA added to trust store")




}
