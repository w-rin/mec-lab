package gitserver

import (
	"fmt"
	"github.com/epmd-edp/codebase-operator/v2/pkg/gerrit"
	"github.com/epmd-edp/codebase-operator/v2/pkg/model"
	"github.com/epmd-edp/codebase-operator/v2/pkg/util"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	goGit "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"gopkg.in/src-d/go-git.v4/storage/memory"
	v1 "k8s.io/api/core/v1"
	coreV1Client "k8s.io/client-go/kubernetes/typed/core/v1"
	"strings"
	"time"
)

type GitSshData struct {
	Host string
	User string
	Key  string
	Port int32
}

func CommitChanges(directory string) error {
	log.Info("Start commiting changes", "directory", directory)
	r, err := git.PlainOpen(directory)
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Add(".")
	if err != nil {
		return err
	}

	array := strings.Split(directory, "/")
	cmsg := fmt.Sprintf("Add template for %v", array[len(array)-1])
	_, err = w.Commit(cmsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "admin",
			Email: "admin@epam-edp.com",
			When:  time.Now(),
		},
	})
	if err != nil {
		return err
	}
	log.Info("Changes have been commited", "directory", directory)
	return nil
}

func PushChanges(key, user, directory string) error {
	log.Info("Start pushing changes", "directory", directory)
	auth, err := initAuth(key, user)
	if err != nil {
		return err
	}

	r, err := git.PlainOpen(directory)
	if err != nil {
		return err
	}

	err = r.Push(&git.PushOptions{
		RemoteName: "origin",
		RefSpecs: []config.RefSpec{
			"refs/heads/*:refs/heads/*",
			"refs/tags/*:refs/tags/*",
		},
		Auth: auth,
	})
	if err != nil {
		return err
	}
	log.Info("Changes has been pushed", "directory", directory)
	return nil
}

func CheckPermissions(repo string, user string, pass string) (accessible bool) {
	log.Info("checking permissions", "user", user, "repository", repo)
	r, _ := git.Init(memory.NewStorage(), nil)
	remote, _ := r.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{repo},
	})
	rfs, err := remote.List(&git.ListOptions{
		Auth: &http.BasicAuth{
			Username: user,
			Password: pass,
		}})
	if err != nil {
		log.Error(err, fmt.Sprintf("User %v do not have access to %v repository", user, repo))
		return false
	}
	return len(rfs) != 0
}

func CloneRepositoryBySsh(key, user, repoUrl, destination string) error {
	log.Info("Start cloning", "repository", repoUrl)
	auth, err := initAuth(key, user)
	if err != nil {
		return err
	}

	_, err = git.PlainClone(destination, false, &git.CloneOptions{
		URL:  repoUrl,
		Auth: auth,
	})
	if err != nil {
		return err
	}
	log.Info("End cloning", "repository", repoUrl)
	return nil
}

func CloneRepository(repo, user, pass, destination string) error {
	log.Info("Start cloning", "repository", repo)
	_, err := git.PlainClone(destination, false, &git.CloneOptions{
		URL: repo,
		Auth: &http.BasicAuth{
			Username: user,
			Password: pass,
		}})
	if err != nil {
		return err
	}
	log.Info("End cloning", "repository", repo)
	return nil
}

func initAuth(key, user string) (*goGit.PublicKeys, error) {
	log.Info("Initializing auth", "user", user)
	signer, err := ssh.ParsePrivateKey([]byte(key))
	if err != nil {
		return nil, err
	}

	return &goGit.PublicKeys{
		User:   user,
		Signer: signer,
		HostKeyCallbackHelper: goGit.HostKeyCallbackHelper{
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		},
	}, nil
}

func checkConnectionToGitServer(c coreV1Client.CoreV1Client, gitServer model.GitServer) (bool, error) {
	log.Info("Start CheckConnectionToGitServer method", "Git host", gitServer.GitHost)

	sshSecret, err := util.GetSecret(c, gitServer.NameSshKeySecret, gitServer.Namespace)
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("an error has occurred  while getting %v secret", gitServer.NameSshKeySecret))
	}

	gitSshData := extractSshData(gitServer, sshSecret)

	log.Info("Data from request is extracted", "host", gitSshData.Host, "port", gitSshData.Port)
 //ここまでうまくいってる

	a := isGitServerAccessible(gitSshData)
	log.Info("Git server", "accessible", a)
	return a, nil
}

func isGitServerAccessible(data GitSshData) bool {
	log.Info("Start executing IsGitServerAccessible method to check connection to server", "host", data.Host)
	//これ自体もうまくいってる
	sshClient, err := sshInitFromSecret(data)
	if err != nil {
		log.Info(fmt.Sprintf("An error has occurred while initing SSH client. Check data in Git Server resource and secret: %v", err))
		return false
	}

	var s *ssh.Session
	var c *ssh.Client
	if s, c, err = sshClient.NewSession(); err != nil {
		log.Info(fmt.Sprintf("An error has occurred while connecting to server. Check data in Git Server resource and secret: %v", err))
		return false
	}
	defer s.Close()
	defer c.Close()

	return s != nil && c != nil
}

func extractSshData(gitServer model.GitServer, secret *v1.Secret) GitSshData {
	return GitSshData{
		Host: gitServer.GitHost,
		User: gitServer.GitUser,
		Key:  string(secret.Data[util.PrivateSShKeyName]),
		Port: gitServer.SshPort,
	}
}

func sshInitFromSecret(data GitSshData) (gerrit.SSHClient, error) {
	sshConfig := &ssh.ClientConfig{
		User: data.User,
		Auth: []ssh.AuthMethod{
			publicKey(data.Key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client := &gerrit.SSHClient{
		Config: sshConfig,
		Host:   data.Host,
		Port:   data.Port,
	}
	log.Info("SSH Client has been initialized: Host: %v Port: %v", data.Host, data.Port)
	return *client, nil
	//ここではもうエラーが出ている
}

func publicKey(key string) ssh.AuthMethod {
  signer, err := ssh.ParsePrivateKey([]byte(key))
	if err != nil {
		panic(err)
	}
	return ssh.PublicKeys(signer)
}
