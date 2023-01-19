package golink

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/jlaffaye/ftp"
	"github.com/link1st/go-stress-testing/helper"
	"github.com/link1st/go-stress-testing/model"
)

var logErr = log.New(os.Stderr, "", 0)

func operate(c *ftp.ServerConn, fileName string) (int64, error) {
	var (
		err error = nil
		buf []byte
	)

	r, err := c.Retr(fileName)
	if err != nil {
		logErr.Println("Failed to get file", err)
		return 0, err
	}
	defer r.Close()

	if buf, err = ioutil.ReadAll(r); err != nil {
		logErr.Println("Failed to read", err)
		return 0, err
	}
	return int64(len(buf)), err
}

func ftpsLogin(request *model.Request) (c *ftp.ServerConn, err error) {
	conf := &tls.Config{
		InsecureSkipVerify: true,
	}
	c, err = ftp.Dial(request.URL, ftp.DialWithTLS(conf))
	if err != nil {
		logErr.Println("Failed to Dial:", err)
		return nil, err
	}

	err = c.Login("admin", "admin")
	if err != nil {
		logErr.Println("Failed to login:", err)
		return nil, err
	}
	return
}

func ftpsRun(c *ftp.ServerConn, request *model.Request) (lens int64, err error) {

	// Do something with the FTP conn
	lens, err = operate(c, request.Body)
	if err != nil {
		return lens, err
	}

	return lens, err
}

// Ftps request
func Ftps(ctx context.Context, chanID uint64, ch chan<- *model.RequestResults, totalNumber uint64, wg *sync.WaitGroup,
	request *model.Request) {
	var (
		startTime   time.Time
		requestTime uint64
		lens        int64 = 0
		err         error
	)
	defer func() {
		wg.Done()
	}()
	c, err := ftpsLogin(request)
	if err != nil {
		return
	}
	for i := uint64(0); i < totalNumber; i++ {
		startTime = time.Now()
		lens, err = ftpsRun(c, request)
		requestTime = uint64(helper.DiffNano(startTime))

		rr := &model.RequestResults{
			Time:          requestTime,
			IsSucceed:     err == nil,
			ReceivedBytes: lens,
		}
		rr.SetID(chanID, i)
		ch <- rr
	}
	if err = c.Quit(); err != nil {
		logErr.Println("Failed to quit:", err)
	}
}
