package qiniu

import (
	//"fmt"
	"io/ioutil"
	"os"
	//"strconv"
	//"testing"
	//
	//storagedriver "github.com/docker/distribution/registry/storage/driver"
	//"github.com/docker/distribution/registry/storage/driver/testsuites"
	//"gopkg.in/check.v1"
	"testing"
	"fmt"
)

var (
	d *Driver
)

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

}

func TestDriver_Name(t *testing.T) {

	if d.Name() != "qiniu" {
		t.Error("Invalid name")
	}
}
//
//func TestDriver_GetContent(t *testing.T) {
//	data,err := d.GetContent(nil, "/docker/test")
//	if err != nil {
//		t.Error(err)
//	}
//	fmt.Print(string(data))
//}
//
//func TestDriver_PutContent(t *testing.T) {
//	err := d.PutContent(nil, "/docker/test2", []byte("abc"))
//	if err != nil {
//		t.Error(err)
//	}
//}
//
//func TestDriver_List(t *testing.T) {
//	items, err := d.List(nil, "/docker");
//	if err != nil {
//		t.Error(err)
//	}
//	fmt.Print(items)
//}
//
//func TestDriver_Stat(t *testing.T) {
//	info, err := d.Stat(nil, "/docker/test")
//	if err != nil {
//		t.Error(err)
//	}
//	fmt.Print(info)
//}

//func TestDriver_URLFor(t *testing.T) {
//	url, err := d.URLFor(nil, "/docker/test", nil)
//	if err != nil {
//		t.Error(err)
//	}
//	fmt.Print("url="+url)
//}
//
//func TestDriver_Move(t *testing.T) {
//	err := d.Move(nil,"/docker/test2", "/docker/test3")
//	if err != nil {
//		t.Error(err)
//	}
//}
//
//func TestDriver_Delete(t *testing.T) {
//	err := d.Delete(nil, "/docker/test3")
//	if err != nil {
//		t.Error(err)
//	}
//}

//func TestDriver_Reader(t *testing.T) {
//	reader, err := d.Reader(nil,"/docker/test",1);
//	if err != nil {
//		t.Error(err)
//	}
//	var p []byte
//	reader.Read(p)
//	fmt.Print(len(p))
//	fmt.Print(string(p))
//}

func TestDriver_Writer(t *testing.T) {

	writer, err := d.Writer(nil,"/docker/testWriter",false)
	if err != nil {
		t.Error(err)
	}

	length, err := writer.Write([]byte("testWriter"))
	if err != nil {
		t.Error(err)
	}
	err = writer.Commit()
	if err != nil {
		t.Error(err)
	}
	err = writer.Close()
	if err != nil {
		t.Error(err)
	}
	fmt.Print(length)
}

//func Test(t *testing.T) { check.TestingT(t) }
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
//	parameters := map[string]interface{}{}
//	parameters[paramAccountKey] = secretKey
//	parameters[paramAccountName] = accessKey
//	parameters[paramBucket] = bucket
//	parameters[paramZone] = zone
//	parameters[paramDomain] = domain
//	parameters[paramIsPrivate] = isPrivate
//
//
//	d, _ = FromParameters(parameters)
//
//	qiniuDriverConstructor = func(rootDirectory string) (*Driver, error) {
//		var err error
//		zoneInt := 0
//		if zone != "" {
//			zoneInt, err = strconv.Atoi(zone)
//			if err != nil {
//				return nil, err
//			}
//		}
//
//		isPrivateBool := false
//		if isPrivate != "" {
//			isPrivateBool, err = strconv.ParseBool(isPrivate)
//			if err != nil {
//				return nil, err
//			}
//		}
//
//		parameters := DriverParameters{
//			AccessKey: accessKey,
//			SecretKey: secretKey,
//			Bucket:    bucket,
//			Zone:      zoneInt,
//			IsPrivate: isPrivateBool,
//		}
//
//		return New(parameters)
//	}
//
//	// Skip OSS storage driver tests if environment variable parameters are not provided
//	skipCheck = func() string {
//		if accessKey == "" || secretKey == "" || zone == "" || bucket == "" || isPrivate == "" || bucket == "" {
//			return "Must set QINIU_ACCOUNT_NAME, QINIU_ACCOUNT_KEY, QINIU_BUCKET, QINIU_ZONE, and QINIU_ISPRIVATE to run qiniu tests"
//		}
//		return ""
//	}
//
//	testsuites.RegisterSuite(func() (storagedriver.StorageDriver, error) {
//		return qiniuDriverConstructor(root)
//	}, skipCheck)
//
//}
//
//func TestDriver_Name(t *testing.T) {
//	fmt.Print(d.Name())
//}
