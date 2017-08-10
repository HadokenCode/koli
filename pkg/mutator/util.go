package mutator

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	platform "kolihub.io/koli/pkg/apis/v1alpha1"
	"kolihub.io/koli/pkg/request"

	jwt "github.com/dgrijalva/jwt-go"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
)

var (
	extensionsCodec = api.Codecs.LegacyCodec(v1beta1.SchemeGroupVersion)
)

// Config is the daemon base configuration
type Config struct {
	Host            string `envconfig:"KUBERNETES_SERVICE_HOST" required:"true"`
	TLSInsecure     bool
	TLSServerConfig rest.TLSClientConfig
	TLSClientConfig rest.TLSClientConfig
	Serve           string
	AllowedImages   string
	RegistryImages  string
	KongAPIHost     string
}

// GetServeAddress return the address to bind the server
func (c *Config) GetServeAddress() (string, bool) {
	if len(c.TLSServerConfig.CertFile) > 0 && len(c.TLSServerConfig.KeyFile) > 0 && len(c.Serve) == 0 {
		return ":8443", true
	}
	if len(c.Serve) == 0 {
		return ":8080", false
	}
	return c.Serve, false
}

// GetImages returns of allowed images with the registry as prefix
func (c *Config) GetImages() []string {
	images := []string{}
	for _, img := range strings.Split(c.AllowedImages, ",") {
		images = append(images, filepath.Join(c.RegistryImages, img))
	}
	return images
}

func forbiddenAccessMessage(u *platform.User, customer, org string) string {
	msg := fmt.Sprintf("Permission denied. The user belongs to the customer '%s' and organization '%s', ", u.Customer, u.Organization)
	msg = msg + fmt.Sprintf("but the request was sent to the customer '%s' and organization '%s'. ", customer, org)
	return msg + fmt.Sprintf("Valid values are '[name]-%s-%s'", u.Customer, u.Organization)
}

// decodeJwtToken decodes a jwt token into an UserMeta struct
func decodeJwtToken(header http.Header) (*platform.User, string, error) {
	// [0] = "bearer" / [1] = "<token>"{
	authorization := strings.Split(header.Get("Authorization"), " ")
	if len(authorization) != 2 {
		return nil, "", fmt.Errorf("missing token or bearer in Authorization")
	}
	parts := strings.Split(authorization[1], ".")
	if len(parts) != 3 {
		return nil, "", fmt.Errorf("it's not a valid jwt token")
	}
	// Don't care about validating tokens, only about the token data.
	seg, err := jwt.DecodeSegment(parts[1])
	if err != nil {
		return nil, "", fmt.Errorf("failed decoding segment: %s", err)
	}
	user := &platform.User{}
	return user, authorization[1], json.Unmarshal(seg, user)
}

// getKubernetesUserClients returns clients to interact with the api server
func getKubernetesUserClients(mutatorCfg *Config, bearerToken string) (*kubernetes.Clientset, rest.Interface, error) {
	var config *rest.Config
	var err error
	if mutatorCfg == nil || len(mutatorCfg.Host) == 0 {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, nil, err
		}
	} else {
		config = &rest.Config{Host: mutatorCfg.Host}
		config.TLSClientConfig = mutatorCfg.TLSClientConfig
		config.Insecure = mutatorCfg.TLSInsecure
	}
	config.BearerToken = bearerToken

	var tprConfig *rest.Config
	tprConfig = config
	tprConfig.APIPath = "/apis"
	tprConfig.GroupVersion = &platform.SchemeGroupVersion
	tprConfig.ContentType = runtime.ContentTypeJSON
	tprConfig.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: api.Codecs}
	metav1.AddToGroupVersion(api.Scheme, platform.SchemeGroupVersion)
	platform.SchemeBuilder.AddToScheme(api.Scheme)

	userTprClient, err := rest.RESTClientFor(tprConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed retrieving user k8s tpr client from config [%v]", err)
	}
	userKubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed retrieving user k8s client from config [%v]", err)
	}
	return userKubeClient, userTprClient, nil
}

func initializeMetadata(o *metav1.ObjectMeta) {
	if o.Labels == nil {
		o.Labels = make(map[string]string)
	}
	if o.Annotations == nil {
		o.Annotations = make(map[string]string)
	}
}

func NewKongClient(client request.HTTPClient, apiURL string) (request.Interface, error) {
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}
	return request.NewRequest(client, u), nil
}