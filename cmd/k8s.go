package cmd

import (
	"context"
	"fmt"
	"github.com/pytimer/k8sutil/apply"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
	// "github.com/antelman107/net-wait-go/wait"
	// "io/fs"

	// corev1 "k8s.io/api/core/v1"
	"io/ioutil"
	// "log"
	// "path/filepath"
	"embed"
	// imagepolicyv1alpha1 "k8s.io/api/imagepolicy/v1alpha1"
	// utilruntime "k8s.io/apimachinery/pkg/util/runtime"

	// "k8s.io/client-go/util/homedir"
	"crypto/tls"
	"crypto/x509"
	"github.com/avast/retry-go/v4"
	"gopkg.in/yaml.v3"
	"net/http"
	"path"
	"github.com/apex/log/handlers/cli"
	"github.com/apex/log"
)
func init() {
	log.SetLevel(log.DebugLevel)
	log.SetHandler(cli.Default)
}
//go:embed manifests/*
var folder embed.FS

// var yamlFile []byte

type RetriableError struct {
	Err error
}

var _ error = (*RetriableError)(nil)

func serverReady(kubeconfig string) {
	ctx := log.WithFields(log.Fields{
		"kubeconfig": kubeconfig,
	})
	ctx.Info("Wait for server readiness")

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		ctx.Fatalf("error: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		ctx.Fatalf("error: %v", err)

	}

	for {
		var ready bool
		nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
			LabelSelector: "node-role.kubernetes.io/master=true",
		})
		if err != nil {
			ctx.Fatalf("error: %v", err)

		}
		for _, item := range nodes.Items {
			for _, condition := range item.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == "True" {
					ready = true
				}
			}
		}
		if ready {
			break
		}
	}

}

func waitNodes(kubeconfig string) {
	ctx := log.WithFields(log.Fields{
		"kubeconfig": kubeconfig,
	})
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		ctx.Fatalf("error: %v", err)

	}
	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		ctx.Fatalf("error: %v", err)

	}

	fmt.Println("wait for nodes")
	for {
		var ready []string
		nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			ctx.Fatalf("error: %v", err)

		}
		for _, item := range nodes.Items {
			for _, condition := range item.Status.Conditions {
				if condition.Type == "Ready" && condition.Status == "True" {
					ready = append(ready, item.Name)
				}
			}
		}
		if len(ready) == 2 {
			break
		}
	}
	fmt.Println("nodes UP")
}

type KubeConfigYML struct {
	APIVersion string `yaml:"apiVersion"`
	Clusters   []struct {
		Cluster struct {
			CertificateAuthorityData string `yaml:"certificate-authority-data"`
			Server                   string `yaml:"server"`
		} `yaml:"cluster"`
		Name string `yaml:"name"`
	} `yaml:"clusters"`
	Contexts []struct {
		Context struct {
			Cluster string `yaml:"cluster"`
			User    string `yaml:"user"`
		} `yaml:"context"`
		Name string `yaml:"name"`
	} `yaml:"contexts"`
	CurrentContext string `yaml:"current-context"`
	Kind           string `yaml:"kind"`
	Preferences    struct {
	} `yaml:"preferences"`
	Users []struct {
		Name string `yaml:"name"`
		User struct {
			ClientCertificateData string `yaml:"client-certificate-data"`
			ClientKeyData         string `yaml:"client-key-data"`
		} `yaml:"user"`
	} `yaml:"users"`
}

func healthAPI(kubeconfig string) {
	ctx := log.WithFields(log.Fields{
		"kubeconfig": kubeconfig,
	})

			yamlFile, err := ioutil.ReadFile(kubeconfig)
			if err != nil {
				ctx.Fatalf("error reading file",err) 
			}
		

	kubeconfigfile := KubeConfigYML{}
	err = yaml.Unmarshal([]byte(yamlFile), &kubeconfigfile)
	if err != nil {
		ctx.Fatalf("error: %v", err)
	}
	CertificateAuthorityData := decodeB64(kubeconfigfile.Clusters[0].Cluster.CertificateAuthorityData)
	ClientCertificateData := decodeB64(kubeconfigfile.Users[0].User.ClientCertificateData)
	ClientKeyData := decodeB64(kubeconfigfile.Users[0].User.ClientKeyData)

	cert, err := tls.X509KeyPair([]byte(ClientCertificateData), []byte(ClientKeyData))
	if err != nil {
		fmt.Println(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM([]byte(CertificateAuthorityData))

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	tlsConfig.BuildNameToCertificate()
	transport := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: transport}

	url := "https://127.0.0.1:6443/healthz"
	var body []byte

	retry.Do(
		func() error {
			resp, err := client.Get(url)
			if err != nil {
				return err
			}
			defer resp.Body.Close()
			body, err = ioutil.ReadAll(resp.Body)
			if err != nil {
				return err
			}
			if string(body) != "ok" {

				return &RetriableError{
					Err: err,
				}
			}else{
				ctx.Debug("API healthy")
			}

			return nil
		},
	)
}

func getAllFilenames(fs embed.FS, dir string) (out []string, err error) {
	ctx := log.WithFields(log.Fields{
		"dir": dir,
	})
	if len(dir) == 0 {
		dir = "."
	}

	entries, err := fs.ReadDir(dir)
	if err != nil {
		// return nil, err
		ctx.Fatalf("error: %v", err)

	}

	for _, entry := range entries {
		fp := path.Join(dir, entry.Name())
		if entry.IsDir() {
			res, err := getAllFilenames(fs, fp)
			if err != nil {
				// return nil, err
				ctx.Fatalf("error: %v", err)

			}

			out = append(out, res...)

			continue
		}

		out = append(out, fp)
	}

	return out, err
}

func applyManifests(kubeconfig string, wildcard string) {
	ctx := log.WithFields(log.Fields{
		"domain": wildcard,

	})
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		ctx.Fatalf("error: %v", err)

	}


	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		ctx.Fatalf("error: %v", err)
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		ctx.Fatalf("error: %v", err)
	}

	fileList, err := getAllFilenames(folder, "manifests")
	if err != nil {
		ctx.Fatalf("error: %v", err)
	}
	applyOptions := apply.NewApplyOptions(dynamicClient, discoveryClient)

	for _, filename := range fileList {

		file, _ := folder.ReadFile(filename)

		templatedYaml := strings.Replace(string(file), "${DOMAIN}", wildcard, -1)
		if err := applyOptions.Apply(context.TODO(), []byte(templatedYaml)); err != nil {
			ctx.Fatalf("apply error: %v", err)
		}
	}


}

func tlsSecret(kubeconfig string) {
	ctx := log.WithFields(log.Fields{
		"kubeconfig": kubeconfig,
	})
	key, err := ioutil.ReadFile(UserHomeDir() + "/.k3dev/certificates/key.pem")
	cert, err := ioutil.ReadFile(UserHomeDir() + "/.k3dev/certificates/cert.pem")

	if err != nil {
		ctx.Fatalf("error: %v", err)
	}

	// data:="test"
	secret := fmt.Sprintf(`
apiVersion: v1
data:
  tls.crt: %s
  tls.key: %s
kind: Secret
metadata:
  name: ssl
type: kubernetes.io/tls
  `, encodeB64(string(cert)), encodeB64(string(key)))

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Println(err)
		panic(err.Error())
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		ctx.Fatalf("error: %v", err)
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		ctx.Fatalf("error: %v", err)
	}
	applyOptions := apply.NewApplyOptions(dynamicClient, discoveryClient)

	if err := applyOptions.Apply(context.TODO(), []byte(secret)); err != nil {
		ctx.Fatalf("apply error: %v", err)
	}

}



