package plugin

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sethvargo/go-password/password"
	"golang.org/x/crypto/bcrypt"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	apis "github.com/Danil-Grigorev/rancher-bind/pkg/apis"
	managementv3 "github.com/Danil-Grigorev/rancher-bind/pkg/apis/rancher/management/v3"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const commonName = "rancher-bind"

func GetServer(ctx context.Context, cl client.Client) (string, error) {
	serverUrl := &managementv3.Setting{ObjectMeta: metav1.ObjectMeta{
		Name: "server-url",
	}}
	if err := cl.Get(ctx, client.ObjectKeyFromObject(serverUrl), serverUrl); err != nil {
		return "", err
	}

	return serverUrl.Value, nil
}

func HashPasswordString(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("problem encrypting password: %w", err)
	}

	return string(hash), nil
}

func GenerateRandomPassword() (string, string, error) {
	password, err := password.Generate(64, 10, 10, false, false)
	if err != nil {
		return "", "", fmt.Errorf("problem generating password: %w", err)
	}

	hash, err := HashPasswordString(password)
	return password, hash, err
}

func createOrUpdate(ctx context.Context, cl client.Client, obj client.Object) error {
	err := cl.Create(ctx, obj)
	if err == nil {
		return nil
	}

	if apierrors.IsAlreadyExists(err) {
		if err := cl.Update(ctx, user); err != nil {
			return fmt.Errorf("Unable to update existing object: %w", err)
		}

		return nil
	}

	return nil
}

func CreateUser(ctx context.Context, cl client.Client, passwordHash string) (*managementv3.User, error) {
	user := &managementv3.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: commonName,
		},
		Username: commonName,
		Password: passwordHash,
	}

	if err := createOrUpdate(ctx, cl, user); err == nil {
		return nil, fmt.Errorf("Unable to create a new user: %w", err)
	}

	return user, nil
}

func ResetPassword(ctx context.Context, cl client.Client) error {
	user := &managementv3.User{
		ObjectMeta: metav1.ObjectMeta{
			Name: commonName,
		},
		Username: commonName,
		Password: "",
	}

	if err := cl.Update(ctx, user); err != nil {
		return fmt.Errorf("Unable to reset user password: %w", err)
	}

	return nil
}

func CreateClusterRole(ctx context.Context, cl client.Client, user *managementv3.User) (*managementv3.GlobalRole, error) {
	role := &managementv3.GlobalRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: user.Name,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{managementv3.GroupVersion.Group},
				Resources: []string{"clusters"},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
			},
			{
				APIGroups: []string{"provisioning.cattle.io"},
				Resources: []string{"clusters"},
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
			},
		},
	}

	if err := createOrUpdate(ctx, cl, role); err == nil {
		return nil, fmt.Errorf("Unable to create a new role: %w", err)
	}

	return role, nil
}

func CreateRoleBinding(ctx context.Context, cl client.Client, user *managementv3.User) (*managementv3.GlobalRoleBinding, error) {
	binding := &managementv3.GlobalRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: user.Name,
		},
		GlobalRoleName: user.Name,
		UserName:       user.Name,
	}

	if err := createOrUpdate(ctx, cl, binding); err == nil {
		return nil, fmt.Errorf("Unable to create a new role binding: %w", err)
	}

	return binding, nil
}

func AuthenticateUser(serverUrl string, requestBody *apis.Login) (*apis.LoginResponse, error) {
	loginURL := serverUrl + "/v3-public/localProviders/local?action=login"
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}

	requestDataJSON, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("Error marshaling login data: %w", err)
	}

	resp, err := client.Post(loginURL, "application/json", bytes.NewBuffer(requestDataJSON))
	if err != nil {
		return nil, fmt.Errorf("downloading token: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading token: %w", err)
	}

	response := &apis.LoginResponse{}
	if err := json.Unmarshal(data, response); err != nil {
		return nil, fmt.Errorf("Error parsing the login response: %w", err)
	}

	return response, nil
}

func prepare(req *http.Request, token string) {
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(token)))
	req.Header.Set("Content-Type", "application/json")
}

func CollectKubeconfig(serverUrl, token string) (*apis.ConfigResponse, error) {
	kubeconfigURL := serverUrl + "/v3/clusters/local?action=generateKubeconfig"
	client := &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}}

	req, err := http.NewRequest("POST", kubeconfigURL, nil)
	if err != nil {
		return nil, err
	}

	prepare(req, token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("downloading kubeconfig: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading kubeconfig: %w", err)
	}

	response := &apis.ConfigResponse{}
	if err := json.Unmarshal(data, response); err != nil {
		return nil, fmt.Errorf("Error parsing the kubeconfig response: %w", err)
	}

	return response, nil
}
