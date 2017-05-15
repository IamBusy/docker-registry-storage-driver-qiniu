package qiniu

import (
	"os"
	"io/ioutil"
	"fmt"
	"testing"
	"github.com/docker/distribution/registry/storage/driver/testsuites"
	storagedriver "github.com/docker/distribution/registry/storage/driver"
	"gopkg.in/check.v1"
)

var (
	d *Driver
)

func Test(t *testing.T) { check.TestingT(t) }

var qiniuDriverConstructor func(rootDirectory string) (*Driver, error)

func init() {
	accessKey := os.Getenv("QINIU_ACCOUNT_NAME")
	secretKey := os.Getenv("QINIU_ACCOUNT_KEY")
	bucket := os.Getenv("QINIU_BUCKET")
	zone := os.Getenv("QINIU_ZONE")
	domain := os.Getenv("QINIU_DOMAIN")
	isPrivate := os.Getenv("QINIU_ISPRIVATE")
	root, err := ioutil.TempDir("", "driver-")
	if err != nil {
		panic(err)
	}
	defer os.Remove(root)

	parameters := map[string]interface{}{}
	parameters[paramAccountKey] = secretKey
	parameters[paramAccountName] = accessKey
	parameters[paramBucket] = bucket
	parameters[paramZone] = zone
	parameters[paramDomain] = domain
	parameters[paramIsPrivate] = isPrivate

	d, _ = FromParameters(parameters)

	qiniuDriverConstructor = func(rootDirectory string) (*Driver, error) {

		parameters := DriverParameters{
			AccessKey:     accessKey,
			SecretKey:     secretKey,
			Bucket:        bucket,
			Zone:          zone,
			IsPrivate:     isPrivate,
		}

		return New(parameters)
	}

	// Skip OSS storage driver tests if environment variable parameters are not provided
	skipCheck = func() string {
		if accessKey == "" || secretKey == "" || zone == "" || bucket == "" || isPrivate == "" || bucket == "" {
			return "Must set QINIU_ACCOUNT_NAME, QINIU_ACCOUNT_KEY, QINIU_BUCKET, QINIU_ZONE, and QINIU_ISPRIVATE to run qiniu tests"
		}
		return ""
	}

	testsuites.RegisterSuite(func() (storagedriver.StorageDriver, error) {
		return qiniuDriverConstructor(root)
	}, skipCheck)

}

func TestDriver_Name(t *testing.T) {
	fmt.Print(d.Name())
}

//
//// Hook up gocheck into the "go test" runner.
//func Test(t *testing.T) { check.TestingT(t) }
//
//
//var qiniuDriverConstructor func(rootDirectory string) (*Driver, error)
//
//var skipCheck func() string
//
//func init() {
//	accessKey := os.Getenv("QINIU_ACCOUNT_NAME")
//	secretKey := os.Getenv("QINIU_ACCOUNT_KEY")
//	bucket := os.Getenv("QINIU_BUCKET")
//	zone := os.Getenv("QINIU_ZONE")
//	domain := os.Getenv("QINIU_DOMAIN")
//	isPrivate := os.Getenv("QINIU_ISPRIVATE")
//	root, err := ioutil.TempDir("", "driver-")
//	if err != nil {
//		panic(err)
//	}
//	defer os.Remove(root)
//
//	qiniuDriverConstructor = func(rootDirectory string) (*Driver, error) {
//
//		parameters := map[string]interface{}{}
//		parameters[paramAccountKey] = secretKey
//		parameters[paramAccountName] = accessKey
//		parameters[paramBucket] = bucket
//		parameters[paramZone] = zone
//		parameters[paramDomain] = domain
//		parameters[paramIsPrivate] = isPrivate
//
//		return FromParameters(parameters)
//	}
//
//	// Skip qiniu storage driver tests if environment variable parameters are not provided
//	skipCheck = func() string {
//		if accessKey == "" || secretKey == ""  || bucket == "" {
//			return "Must set QINIU_ACCOUNT_NAME, QINIU_ACCOUNT_KEY, QINIU_BUCKET and QINIU_ZONE to run qiniu tests"
//		}
//		return ""
//	}
//
//	testsuites.RegisterSuite(func() (storagedriver.StorageDriver, error) {
//		return qiniuDriverConstructor(root)
//	}, skipCheck)
//}
