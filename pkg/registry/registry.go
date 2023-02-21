package registry

import (
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/google/go-containerregistry/pkg/registry"
)

type portOrError struct {
	port *int
	err  error
}

// func ServeViaRuntime(via string) (*string, error) {
// 	ps, err := exec.Command(via, "ps", "-f", "name=no-registry", "--format", "{{.Ports}}").Output()
// 	if err != nil {
// 		return nil, err
// 	}

// 	var address string

// 	if len(ps) == 0 {
// 		fmt.Println("registry doesn't exist")
// 		listen, err := net.Listen("tcp", "0.0.0.0:4100")
// 		if err != nil {
// 			listen, err = net.Listen("tcp", "0.0.0.0:0")
// 			if err != nil {
// 				return nil, err
// 			}
// 		}
// 		port := listen.Addr().(*net.TCPAddr).Port
// 		err = listen.Close()
// 		if err != nil {
// 			return nil, err
// 		}
// 		create := exec.Command(via, "container", "run", "-dt", "--rm", "-p", fmt.Sprintf("%d:5000", port), "--name", "no-registry", "docker.io/library/registry:2")
// 		if err = create.Run(); err != nil {
// 			return nil, err
// 		}
// 		address = fmt.Sprintf("0.0.0.0:%d", port)
// 	} else {
// 		split := strings.Split(string(ps), "->")
// 		address = split[0]
// 	}

// 	fmt.Println(address)

// 	return nil, err
// }

func Serve() (*int, error) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	result := make(chan portOrError, 1)

	go func() {
		listen, err := net.Listen("tcp", "0.0.0.0:4100")
		if err != nil {
			listen, err = net.Listen("tcp", "0.0.0.0:0")
			if err != nil {
				result <- portOrError{err: err}
				return
			}
		}
		port := listen.Addr().(*net.TCPAddr).Port
		result <- portOrError{port: &port}
		err = http.Serve(listen, registry.New(registry.Logger(log.New(ioutil.Discard, "", log.LstdFlags))))
		if err != nil {
			result <- portOrError{err: err}
		}

	}()

	res := <-result
	return res.port, res.err
}
