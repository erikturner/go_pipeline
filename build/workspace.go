package build

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"awesomeProject/git"
	"time"
)

type doneChannel chan struct{}

type workOrder struct {
	w                 io.Writer
	repo              string
	pkg               string
	branch            string
	environment       string
	baseDir           string // The working directory
	workspaceDir      string // The working directory for this work order (the gopath)
	srcDir            string // The source directory
	gitDir            string // Where the git repo for this work order lives
	installDir        string // Where to install the build targets
	buildNumber       string
	buildTargets      map[string]string
	cleanup           []string
	targetBoxes       map[string][]string
	done              doneChannel
	failed            bool
	err               error
	username          string
	submitTime        time.Time
	waitDuration      time.Duration
	executeStartTime  time.Time
	executionDuration time.Duration
	commitInfo        []git.CommitInfo
}

func getSource(wo *workOrder) error {

	err := prepareWorkspace(wo)
	if err != nil {
		return err
	}

	if !git.IsRepo(wo.gitDir) {

		/*-----------------------------------------------------------------
		|
		| Remove the directory if it already exists
		|
		*-----------------------------------------------------------------*/
		if _, err := os.Stat(wo.gitDir); err == nil {
			err = respond(wo.w, fmt.Sprintf("[%s] Removing git repository directory [%s].", wo.pkg, wo.gitDir))
			if err != nil {
				log.Println(err)
				return err
			}
			if err = os.RemoveAll(wo.gitDir); err != nil {
				err = fmt.Errorf("[%s] Error removing git repository directory [%s]:\n%v", wo.pkg, wo.gitDir, err)
				log.Println(err)
				return err
			}
		}

		/*-----------------------------------------------------------------
		|
		| Clone the repository
		|
		*-----------------------------------------------------------------*/
		err := respond(wo.w, fmt.Sprintf("[%s] Cloning git repository [%s] into directory [%s].", wo.pkg, wo.repo, wo.gitDir))
		if err != nil {
			log.Println(err)
			return err
		}
		if err = git.Clone(wo.repo, wo.gitDir); err != nil {
			err = fmt.Errorf("[%s] Error cloning repository [%s] into directory [%s]:\n%v", wo.pkg, wo.repo, wo.gitDir, err)
			log.Println(err)
			return err
		}
	}

	/*-----------------------------------------------------------------
	|
	| Fetch the repository
	|
	*-----------------------------------------------------------------*/
	if err := respond(wo.w, fmt.Sprintf("[%s] Fetching source code in directory [%s].", wo.pkg, wo.gitDir)); err != nil {
		log.Println(err)
		return err
	}
	if err = git.Fetch(wo.gitDir); err != nil {
		err = fmt.Errorf("[%s] Error fetching in directory [%s]:\n%v", wo.pkg, wo.gitDir, err)
		log.Println(err)
		return err
	}

	/*-----------------------------------------------------------------
	|
	| Clean up the repository
	|
	*-----------------------------------------------------------------*/
	if err := respond(wo.w, fmt.Sprintf("[%s] Hard resetting git repository in directory [%s].", wo.pkg, wo.gitDir)); err != nil {
		log.Println(err)
		return err
	}
	if err = git.Reset(wo.gitDir); err != nil {
		err = fmt.Errorf("[%s] Error hard resetting git repository in directory [%s]:\n%v", wo.pkg, wo.gitDir, err)
		log.Println(err)
		return err
	}

	if err := respond(wo.w, fmt.Sprintf("[%s] Cleaning repository in directory [%s].", wo.pkg, wo.gitDir)); err != nil {
		log.Println(err)
		return err
	}
	if err = git.Clean(wo.gitDir); err != nil {
		err = fmt.Errorf("[%s] Error cleaning repository in directory [%s]:\n%v", wo.pkg, wo.gitDir, err)
		log.Println(err)
		return err
	}

	/*-----------------------------------------------------------------
	|
	| Switch to the proper branch
	|
	*-----------------------------------------------------------------*/
	if err := respond(wo.w, fmt.Sprintf("[%s] Checking out branch [%s] in directory [%s].", wo.pkg, wo.branch, wo.gitDir)); err != nil {
		log.Println(err)
		return err
	}
	if err = git.Checkout(wo.gitDir, wo.branch); err != nil {
		err = fmt.Errorf("[%s] Error checking out branch [%s] in directory [%s]:\n%v", wo.pkg, wo.branch, wo.gitDir, err)
		log.Println(err)
		return err
	}

	/*-----------------------------------------------------------------
	|
	| Pull in new changes
	|
	*-----------------------------------------------------------------*/
	if err := respond(wo.w, fmt.Sprintf("[%s] Pulling new changes into branch [%s] in directory [%s].", wo.pkg, wo.branch, wo.gitDir)); err != nil {
		log.Println(err)
		return err
	}
	if err = git.Pull(wo.gitDir); err != nil {
		err = fmt.Errorf("[%s] Error pulling new changes into branch [%s] in directory [%s]:\n%v", wo.pkg, wo.branch, wo.gitDir, err)
		log.Println(err)
		return err
	}

	return nil
}

func prepareWorkspace(wo *workOrder) error {

	/*-----------------------------------------------------------------
	|
	| Create the base directory for the users workspaces if it
	| doesn't exist
	|
	*-----------------------------------------------------------------*/
	if _, err := os.Stat(wo.baseDir); os.IsNotExist(err) {
		err = respond(wo.w, fmt.Sprintf("[%s] Missing base directory [%s]. Creating it now.", wo.pkg, wo.baseDir))
		if err != nil {
			log.Println(err)
			return err
		}
		err = createBaseDir(wo)
		if err != nil {
			err = fmt.Errorf("[%s] Error creating base directory [%s]:\n%v", wo.pkg, wo.baseDir, err)
			log.Println(err)
			return err
		}
	}

	/*-----------------------------------------------------------------
	|
	| Create the directory structure for this git repository if it
	| doesn't already exist
	|
	*-----------------------------------------------------------------*/
	if _, err := os.Stat(wo.gitDir); os.IsNotExist(err) {
		err = respond(wo.w, fmt.Sprintf("[%s] Missing repository directory structure [%s]. Creating it now.", wo.pkg, wo.gitDir))
		if err != nil {
			log.Println(err)
			return err
		}
		err = os.MkdirAll(wo.gitDir, 0700)
		if err != nil {
			err = fmt.Errorf("[%s] Error creating repository directory structure [%s]:\n%v", wo.pkg, wo.gitDir, err)
			log.Println(err)
			return err
		}
	}

	/*-----------------------------------------------------------------
	|
	| Remove anything outside the git repoitory directory so that it
	| doesn't interfere with the build.
	|
	*-----------------------------------------------------------------*/
	filepath.Walk(wo.workspaceDir, func(path string, info os.FileInfo, err error) error {

		if info.IsDir() {

			if path == wo.gitDir {
				return filepath.SkipDir
			} else {
				return nil
			}

		}

		err = respond(wo.w, fmt.Sprintf("[%s] Removing [%s].", wo.pkg, path))
		if err != nil {
			log.Println(err)
			return err
		}
		if err = os.RemoveAll(path); err != nil {
			err = fmt.Errorf("[%s] Error removing [%s]:\n%v", wo.pkg, path, err)
			log.Println(err)
			return err
		}
		return nil
	})

	return nil
}

var createBaseDirMutex sync.Mutex

func createBaseDir(wo *workOrder) error {

	createBaseDirMutex.Lock()
	defer createBaseDirMutex.Unlock()

	if _, err := os.Stat(wo.baseDir); os.IsNotExist(err) {
		return os.Mkdir(wo.baseDir, 0700)
	}

	return nil
}
